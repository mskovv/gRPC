package grpc_service

import "strings"

type ACLChecker interface {
	CheckAccess(consumer, method string) bool
}

type ACLData map[string][]string

type SimpleACLChecker struct {
	acl ACLData
}

func NewSimpleACLChecker(acl ACLData) *SimpleACLChecker {
	return &SimpleACLChecker{acl: acl}
}

func (a *SimpleACLChecker) CheckAccess(consumer, method string) bool {
	if allowedMethods, ok := a.acl[consumer]; ok {
		for _, allowedMethod := range allowedMethods {
			if allowedMethod == method || strings.Contains(allowedMethod, "*") {
				return true
			}
		}
	}
	return false
}
