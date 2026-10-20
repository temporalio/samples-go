package contextawareencryption

var (
	TenantKeyOrgTenant1      = "org/tenant-1"
	TenantKeyOrgTenant2      = "org/tenant-2"
	TenantKeyOrgTenant3      = "org/tenant-3"
	TenantKeysByOrganization = map[string]string{
		TenantKeyOrgTenant1: "tenant-1-key",
		TenantKeyOrgTenant2: "tenant-2-key",
		TenantKeyOrgTenant3: "tenant-3-key",
	}
)
