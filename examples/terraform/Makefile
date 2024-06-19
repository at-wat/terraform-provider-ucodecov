TARGET             := local-providers/registry.terraform.io/at-wat/ucodecov/99.0.0/linux_amd64/terraform-provider-ucodecov_v99.0.0
TF_CLI_CONFIG_FILE := $(CURDIR)/use-local-provider.rc

export TF_CLI_CONFIG_FILE

test: $(TARGET)
	terraform init
	terraform refresh

$(TARGET): ../../*.go ../../codecov/*.go
	mkdir -p $(dir $@)
	go build -o $@ ../../

clean:
	rm -rf .terraform .terraform.lock.hcl
