package tests

import (
	"testing"
	analyzer "vodka/analyzer"
)

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		params   map[string]interface{}
		expected bool
	}{
		{
			name:     "简单比较",
			expr:     "a > 5",
			params:   map[string]interface{}{"a": 10},
			expected: true,
		},
		{
			name:     "简单比较",
			expr:     "a < 5",
			params:   map[string]interface{}{"a": 10},
			expected: false,
		},
		{
			name:     "逻辑与",
			expr:     "a > 5 && b < 3",
			params:   map[string]interface{}{"a": 10, "b": 2},
			expected: true,
		},
		{
			name:     "逻辑或",
			expr:     "a > 5 || b < 3",
			params:   map[string]interface{}{"a": 4, "b": 4},
			expected: false,
		},
		{
			name:     "逻辑非",
			expr:     "!(a > 5)",
			params:   map[string]interface{}{"a": 4},
			expected: true,
		},
		{
			name:     "复杂表达式",
			expr:     "a < 5 && b > 3 || c == 7",
			params:   map[string]interface{}{"a": 4, "b": 4, "c": 7},
			expected: true,
		},
		{
			name:     "复杂表达式",
			expr:     "a < 5 && (b < 3 || c == 7)",
			params:   map[string]interface{}{"a": 4, "b": 4, "c": 7},
			expected: true,
		},
		{
			name:     "判断null",
			expr:     "a == null",
			params:   map[string]interface{}{"a": nil},
			expected: true,
		},
		{
			name:     "判断非null",
			expr:     "a != null",
			params:   map[string]interface{}{"a": 1},
			expected: true,
		},
		{
			name:     "判断nil",
			expr:     "a == nil",
			params:   map[string]interface{}{"a": nil},
			expected: true,
		},
		{
			name:     "判断非nil",
			expr:     "a != nil",
			params:   map[string]interface{}{"a": 1},
			expected: true,
		},
		// 可以添加更多测试用例
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.EvaluateExpression(tt.expr, tt.params)
			if result != tt.expected {
				t.Errorf("EvaluateExpression(%q, %v) = %v, 期望 %v", tt.expr, tt.params, result, tt.expected)
			}
		})
	}
}
