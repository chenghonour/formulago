/*
 * Copyright 2023 FormulaGo Authors
 *
 * Created by hua
 */

package types

import "testing"

func TestSubStrByLen(t *testing.T) {
	type args struct {
		text  string
		limit int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{"该邮件内容为高度机密信息，各阅读者必须履行已签协议中的保密条款，如有发现违规行为，将会严格按照保密条款中违约责任执行。", 10}, "该邮件内容为高度机密"},
		{"test2", args{"ABCasAVB=2DSDs是的是的", 10}, "ABCasAVB=2"},
		{"test3", args{"审核通过1A", 10}, "审核通过1A"},
		{"test4", args{"1234567890", 5}, "12345"},
		{"test5", args{"ABcdefG", 10}, "ABcdefG"},
		{"test6", args{"过", 1}, "过"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubStrByLen(tt.args.text, tt.args.limit)
			if got != tt.want {
				t.Errorf("SubStrByLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
