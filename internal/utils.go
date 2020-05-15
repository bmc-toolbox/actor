package internal

import "fmt"

func validateParam(params map[string]interface{}, param ...string) error {
	for _, p := range param {
		if _, ok := params[p]; !ok {
			return fmt.Errorf("no required parameter %q", p)
		}
	}
	return nil
}
