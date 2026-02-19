# Commands Used

## Initialize the operator
```bash
operator-sdk init --domain=templarfelix.com --repo=github.com/templarfelix/gameserver-operator
```

## Create Dayz API
```bash
operator-sdk create api --group gameserver --version v1 --kind Dayz --resource --controller
```

## Generate code and manifests
```bash
make generate
make manifests
```

## Build the operator
```bash
make build
```