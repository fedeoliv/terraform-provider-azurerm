package validate

import (
	"fmt"
	"regexp"
)

func OpenShiftMasterPoolName(i interface{}, k string) (warnings []string, errors []error) {
	agentPoolName := i.(string)

	re := regexp.MustCompile(`^[a-z]{1}[a-z0-9]{0,11}$`)
	if re != nil && !re.MatchString(agentPoolName) {
		errors = append(errors, fmt.Errorf("%s must start with a lowercase letter, have max length of 12, and only have characters a-z0-9. Got %q.", k, agentPoolName))
	}

	return warnings, errors
}

func OpenShiftAgentPoolName(i interface{}, k string) (warnings []string, errors []error) {
	agentPoolName := i.(string)

	re := regexp.MustCompile(`^[a-z]{1}[a-z0-9]{0,11}$`)
	if re != nil && !re.MatchString(agentPoolName) {
		errors = append(errors, fmt.Errorf("%s must start with a lowercase letter, have max length of 12, and only have characters a-z0-9. Got %q.", k, agentPoolName))
	}

	return warnings, errors
}
