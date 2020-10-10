.terraform/plugins/registry.terraform.io/at-wat/ucodecov/99.0.0/linux_amd64/terraform-provider-ucodecov_v99.0.0: ../../*.go ../../codecov/*.go
	mkdir -p $(dir $@)
	go build -o $@ ../../

test:
	terraform refresh
