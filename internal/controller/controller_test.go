package controller_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhshin/kube-sentinel/api/v1alpha1"
	"github.com/bhshin/kube-sentinel/internal/controller"
)

func TestReconcileReturnsWithoutErrorForSecurityAgentResource(t *testing.T) {
	t.Parallel()

	reconciler := &controller.SecurityAgentReconciler{}
	agent := &v1alpha1.SecurityAgent{}

	if err := reconciler.Reconcile(agent); err != nil {
		t.Fatalf("Reconcile returned an error for a SecurityAgent resource: %v", err)
	}
}

func TestControllerScopesNamespacedWorkloadOperationsToGlobalTargetNamespace(t *testing.T) {
	t.Parallel()

	files := parseControllerSource(t)
	trackedNamespaceValues := map[string]bool{}
	readsGlobalTargetNamespace := false
	scopesOperationWithGlobalTargetNamespace := false
	scopesWorkloadOperationWithGlobalTargetNamespace := false

	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				for _, rhs := range node.Rhs {
					if expressionContainsGlobalTargetNamespace(rhs, trackedNamespaceValues) {
						readsGlobalTargetNamespace = true
					}
				}

				for index, lhs := range node.Lhs {
					if index >= len(node.Rhs) {
						continue
					}

					rhs := node.Rhs[index]
					if expressionContainsGlobalTargetNamespace(rhs, trackedNamespaceValues) {
						trackAssignedIdentifier(lhs, trackedNamespaceValues)
					}
					if namespaceSelector(lhs) && expressionContainsGlobalTargetNamespace(rhs, trackedNamespaceValues) {
						scopesOperationWithGlobalTargetNamespace = true
					}
				}
			case *ast.ValueSpec:
				for _, value := range node.Values {
					if expressionContainsGlobalTargetNamespace(value, trackedNamespaceValues) {
						readsGlobalTargetNamespace = true
					}
				}

				for index, name := range node.Names {
					if index >= len(node.Values) {
						continue
					}
					if expressionContainsGlobalTargetNamespace(node.Values[index], trackedNamespaceValues) {
						trackedNamespaceValues[name.Name] = true
					}
				}
			case *ast.CallExpr:
				if !isInNamespaceCall(node) {
					if namespacedWorkloadOperationUsesGlobalTargetNamespace(node, trackedNamespaceValues) {
						scopesWorkloadOperationWithGlobalTargetNamespace = true
					}
					break
				}

				for _, argument := range node.Args {
					if expressionContainsGlobalTargetNamespace(argument, trackedNamespaceValues) {
						scopesOperationWithGlobalTargetNamespace = true
					}
				}
			case *ast.CompositeLit:
				for _, element := range node.Elts {
					keyValue, ok := element.(*ast.KeyValueExpr)
					if !ok || !namespaceKey(keyValue.Key) {
						continue
					}
					if expressionContainsGlobalTargetNamespace(keyValue.Value, trackedNamespaceValues) {
						scopesOperationWithGlobalTargetNamespace = true
					}
				}
			}

			return true
		})
	}

	if !readsGlobalTargetNamespace {
		t.Errorf("controller must read spec.global.targetNamespace when reconciling a SecurityAgent")
	}
	if !scopesOperationWithGlobalTargetNamespace {
		t.Errorf("controller must pass spec.global.targetNamespace into namespaced workload operations, such as client.InNamespace(targetNamespace) for list calls or Namespace: targetNamespace for workload objects")
	}
	if !scopesWorkloadOperationWithGlobalTargetNamespace {
		t.Errorf("controller must use spec.global.targetNamespace on a real namespaced workload operation, not only compute or store the namespace value")
	}
}

func parseControllerSource(t *testing.T) []*ast.File {
	t.Helper()

	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list controller files: %v", err)
	}

	fileSet := token.NewFileSet()
	var files []*ast.File
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		t.Fatalf("controller package must contain non-test Go source files")
	}

	return files
}

func expressionContainsGlobalTargetNamespace(expression ast.Expr, trackedNamespaceValues map[string]bool) bool {
	found := false
	ast.Inspect(expression, func(node ast.Node) bool {
		if found {
			return false
		}

		switch node := node.(type) {
		case *ast.SelectorExpr:
			if selectorPath(node) == "Spec.Global.TargetNamespace" || strings.HasSuffix(selectorPath(node), ".Spec.Global.TargetNamespace") {
				found = true
				return false
			}
		case *ast.Ident:
			if trackedNamespaceValues[node.Name] {
				found = true
				return false
			}
		}

		return true
	})

	return found
}

func trackAssignedIdentifier(expression ast.Expr, trackedNamespaceValues map[string]bool) {
	identifier, ok := expression.(*ast.Ident)
	if ok {
		trackedNamespaceValues[identifier.Name] = true
	}
}

func isInNamespaceCall(call *ast.CallExpr) bool {
	path := selectorPath(call.Fun)
	return path == "InNamespace" || strings.HasSuffix(path, ".InNamespace")
}

func namespacedWorkloadOperationUsesGlobalTargetNamespace(call *ast.CallExpr, trackedNamespaceValues map[string]bool) bool {
	if !isWorkloadOperationCall(call) {
		return false
	}

	for _, argument := range call.Args {
		if expressionContainsGlobalTargetNamespace(argument, trackedNamespaceValues) {
			return true
		}
		if inNamespaceArgumentContainsGlobalTargetNamespace(argument, trackedNamespaceValues) {
			return true
		}
		if namespaceCompositeContainsGlobalTargetNamespace(argument, trackedNamespaceValues) {
			return true
		}
	}

	return false
}

func isWorkloadOperationCall(call *ast.CallExpr) bool {
	path := selectorPath(call.Fun)
	for _, operation := range []string{".List", ".Get", ".Create", ".Update", ".Patch", ".Delete", ".DeleteAllOf"} {
		if strings.HasSuffix(path, operation) {
			return true
		}
	}

	return false
}

func inNamespaceArgumentContainsGlobalTargetNamespace(expression ast.Expr, trackedNamespaceValues map[string]bool) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok || !isInNamespaceCall(call) {
		return false
	}

	for _, argument := range call.Args {
		if expressionContainsGlobalTargetNamespace(argument, trackedNamespaceValues) {
			return true
		}
	}

	return false
}

func namespaceCompositeContainsGlobalTargetNamespace(expression ast.Expr, trackedNamespaceValues map[string]bool) bool {
	found := false
	ast.Inspect(expression, func(node ast.Node) bool {
		if found {
			return false
		}

		composite, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}

		for _, element := range composite.Elts {
			keyValue, ok := element.(*ast.KeyValueExpr)
			if !ok || !namespaceKey(keyValue.Key) {
				continue
			}
			if expressionContainsGlobalTargetNamespace(keyValue.Value, trackedNamespaceValues) {
				found = true
				return false
			}
		}

		return true
	})

	return found
}

func namespaceSelector(expression ast.Expr) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "Namespace"
}

func namespaceKey(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == "Namespace"
}

func selectorPath(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name
	case *ast.SelectorExpr:
		prefix := selectorPath(expression.X)
		if prefix == "" {
			return expression.Sel.Name
		}
		return prefix + "." + expression.Sel.Name
	default:
		return ""
	}
}
