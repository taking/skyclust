package network

import (
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"google.golang.org/api/compute/v1"
)

// Pagination, Filtering, and Sorting Helpers

// applyVPCFiltering applies search filtering to VPC list
func applyVPCFiltering(vpcs []VPCInfo, search string) []VPCInfo {
	if search == "" {
		return vpcs
	}

	searchLower := strings.ToLower(search)
	filtered := make([]VPCInfo, 0)

	for _, vpc := range vpcs {
		matches := false
		if strings.Contains(strings.ToLower(vpc.Name), searchLower) {
			matches = true
		}
		if strings.Contains(strings.ToLower(vpc.Description), searchLower) {
			matches = true
		}
		if strings.Contains(strings.ToLower(vpc.ID), searchLower) {
			matches = true
		}
		if matches {
			filtered = append(filtered, vpc)
		}
	}

	return filtered
}

// applyVPCSorting applies sorting to VPC list
func applyVPCSorting(vpcs []VPCInfo, sortBy, sortOrder string) {
	if sortBy == "" {
		return
	}

	sort.Slice(vpcs, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = vpcs[i].Name < vpcs[j].Name
		case "state":
			less = vpcs[i].State < vpcs[j].State
		case "created_at":
			less = vpcs[i].CreationTimestamp < vpcs[j].CreationTimestamp
		default:
			// Default to name sorting
			less = vpcs[i].Name < vpcs[j].Name
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// applyVPCPagination applies pagination to VPC list
func applyVPCPagination(vpcs []VPCInfo, page, limit int) []VPCInfo {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	if offset >= len(vpcs) {
		return []VPCInfo{}
	}

	end := offset + limit
	if end > len(vpcs) {
		end = len(vpcs)
	}

	return vpcs[offset:end]
}

// applySubnetFiltering applies search filtering to Subnet list
func applySubnetFiltering(subnets []SubnetInfo, search string) []SubnetInfo {
	if search == "" {
		return subnets
	}

	searchLower := strings.ToLower(search)
	filtered := make([]SubnetInfo, 0)

	for _, subnet := range subnets {
		matches := false
		if strings.Contains(strings.ToLower(subnet.Name), searchLower) {
			matches = true
		}
		if strings.Contains(strings.ToLower(subnet.Description), searchLower) {
			matches = true
		}
		if strings.Contains(strings.ToLower(subnet.CIDRBlock), searchLower) {
			matches = true
		}
		if strings.Contains(strings.ToLower(subnet.ID), searchLower) {
			matches = true
		}
		if matches {
			filtered = append(filtered, subnet)
		}
	}

	return filtered
}

// applySubnetSorting applies sorting to Subnet list
func applySubnetSorting(subnets []SubnetInfo, sortBy, sortOrder string) {
	if sortBy == "" {
		return
	}

	sort.Slice(subnets, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = subnets[i].Name < subnets[j].Name
		case "state":
			less = subnets[i].State < subnets[j].State
		case "cidr_block":
			less = subnets[i].CIDRBlock < subnets[j].CIDRBlock
		case "created_at":
			less = subnets[i].CreationTimestamp < subnets[j].CreationTimestamp
		default:
			// Default to name sorting
			less = subnets[i].Name < subnets[j].Name
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// applySubnetPagination applies pagination to Subnet list
func applySubnetPagination(subnets []SubnetInfo, page, limit int) []SubnetInfo {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	if offset >= len(subnets) {
		return []SubnetInfo{}
	}

	end := offset + limit
	if end > len(subnets) {
		end = len(subnets)
	}

	return subnets[offset:end]
}

// applySecurityGroupFiltering applies search filtering to Security Group list
func applySecurityGroupFiltering(securityGroups []SecurityGroupInfo, search string) []SecurityGroupInfo {
	if search == "" {
		return securityGroups
	}

	searchLower := strings.ToLower(search)
	filtered := make([]SecurityGroupInfo, 0)

	for _, sg := range securityGroups {
		matches := false
		if strings.Contains(strings.ToLower(sg.Name), searchLower) {
			matches = true
		}
		if strings.Contains(strings.ToLower(sg.Description), searchLower) {
			matches = true
		}
		if strings.Contains(strings.ToLower(sg.ID), searchLower) {
			matches = true
		}
		if matches {
			filtered = append(filtered, sg)
		}
	}

	return filtered
}

// applySecurityGroupSorting applies sorting to Security Group list
func applySecurityGroupSorting(securityGroups []SecurityGroupInfo, sortBy, sortOrder string) {
	if sortBy == "" {
		return
	}

	sort.Slice(securityGroups, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = securityGroups[i].Name < securityGroups[j].Name
		default:
			// Default to name sorting
			less = securityGroups[i].Name < securityGroups[j].Name
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// applySecurityGroupPagination applies pagination to Security Group list
func applySecurityGroupPagination(securityGroups []SecurityGroupInfo, page, limit int) []SecurityGroupInfo {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	if offset >= len(securityGroups) {
		return []SecurityGroupInfo{}
	}

	end := offset + limit
	if end > len(securityGroups) {
		end = len(securityGroups)
	}

	return securityGroups[offset:end]
}

// getTagValue: 태그 목록에서 특정 키의 값을 조회합니다
func (s *Service) getTagValue(tags []ec2Types.Tag, key string) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

// convertTags: 태그 목록을 맵으로 변환합니다
func (s *Service) convertTags(tags []ec2Types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

// parsePort: 포트 범위 문자열에서 포트 번호를 파싱합니다
func (s *Service) parsePort(portRange string) int {
	if portRange == "" {
		return 0
	}
	if strings.Contains(portRange, "-") {
		parts := strings.Split(portRange, "-")
		if len(parts) == 2 {
			if port, err := strconv.Atoi(parts[0]); err == nil {
				return port
			}
		}
	}
	if port, err := strconv.Atoi(portRange); err == nil {
		return port
	}
	return 0
}

// convertGCPFirewallRules: GCP Firewall 규칙을 SecurityGroupRuleInfo로 변환합니다
func (s *Service) convertGCPFirewallRules(firewall *compute.Firewall) []SecurityGroupRuleInfo {
	var rules []SecurityGroupRuleInfo

	// Convert allowed rules
	for _, allowed := range firewall.Allowed {
		for _, portRange := range allowed.Ports {
			rule := SecurityGroupRuleInfo{
				Type:        "ingress",
				Protocol:    allowed.IPProtocol,
				FromPort:    int32(s.parsePort(portRange)),
				ToPort:      int32(s.parsePort(portRange)),
				CIDRBlocks:  firewall.SourceRanges,
				Description: firewall.Description,
			}
			rules = append(rules, rule)
		}
	}

	// Convert denied rules
	for _, denied := range firewall.Denied {
		for _, portRange := range denied.Ports {
			rule := SecurityGroupRuleInfo{
				Type:        "egress",
				Protocol:    denied.IPProtocol,
				FromPort:    int32(s.parsePort(portRange)),
				ToPort:      int32(s.parsePort(portRange)),
				CIDRBlocks:  firewall.DestinationRanges,
				Description: firewall.Description,
			}
			rules = append(rules, rule)
		}
	}

	return rules
}

// convertGCPFirewallRulesFromRequest: CreateSecurityGroupRequest에서 GCP Firewall 규칙을 생성합니다
func (s *Service) convertGCPFirewallRulesFromRequest(req CreateSecurityGroupRequest) []SecurityGroupRuleInfo {
	var rules []SecurityGroupRuleInfo

	for _, port := range req.Ports {
		rule := SecurityGroupRuleInfo{
			Type:        strings.ToLower(req.Direction),
			Protocol:    req.Protocol,
			FromPort:    int32(s.parsePort(port)),
			ToPort:      int32(s.parsePort(port)),
			CIDRBlocks:  req.SourceRanges,
			Description: req.Description,
		}
		rules = append(rules, rule)
	}

	return rules
}

// convertSecurityGroupRules: AWS IP Permission을 SecurityGroupRuleInfo로 변환합니다
func (s *Service) convertSecurityGroupRules(ingress, egress []ec2Types.IpPermission) []SecurityGroupRuleInfo {
	rules := make([]SecurityGroupRuleInfo, 0)

	// Convert ingress rules
	for _, perm := range ingress {
		rule := SecurityGroupRuleInfo{
			Type:         "ingress",
			Protocol:     aws.ToString(perm.IpProtocol),
			FromPort:     aws.ToInt32(perm.FromPort),
			ToPort:       aws.ToInt32(perm.ToPort),
			CIDRBlocks:   make([]string, 0),
			SourceGroups: make([]string, 0),
		}

		// Add CIDR blocks
		for _, ipRange := range perm.IpRanges {
			if ipRange.CidrIp != nil {
				rule.CIDRBlocks = append(rule.CIDRBlocks, aws.ToString(ipRange.CidrIp))
			}
		}

		// Add source groups
		for _, userGroupPair := range perm.UserIdGroupPairs {
			if userGroupPair.GroupId != nil {
				rule.SourceGroups = append(rule.SourceGroups, aws.ToString(userGroupPair.GroupId))
			}
		}

		rules = append(rules, rule)
	}

	// Convert egress rules
	for _, perm := range egress {
		rule := SecurityGroupRuleInfo{
			Type:         "egress",
			Protocol:     aws.ToString(perm.IpProtocol),
			FromPort:     aws.ToInt32(perm.FromPort),
			ToPort:       aws.ToInt32(perm.ToPort),
			CIDRBlocks:   make([]string, 0),
			SourceGroups: make([]string, 0),
		}

		// Add CIDR blocks
		for _, ipRange := range perm.IpRanges {
			if ipRange.CidrIp != nil {
				rule.CIDRBlocks = append(rule.CIDRBlocks, aws.ToString(ipRange.CidrIp))
			}
		}

		// Add source groups
		for _, userGroupPair := range perm.UserIdGroupPairs {
			if userGroupPair.GroupId != nil {
				rule.SourceGroups = append(rule.SourceGroups, aws.ToString(userGroupPair.GroupId))
			}
		}

		rules = append(rules, rule)
	}

	return rules
}
