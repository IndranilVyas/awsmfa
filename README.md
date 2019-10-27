# awsmfa

Small commandline tool to generate AWS Session Credentials for IAM Roles that require MFA.

## Usage
```
Manage your AWS Session Credentials for aws cli/api access IAM Role that has MFA enabled.
awsmfa will generate Session Credentials and save them in default credentials file

Usage:
  awsmfa [flags]

Flags:
  -d, --duration string   Session Duration like 1h, 2h. (default "1h")
  -h, --help              help for awsmfa
  -p, --profile string    profile name found config file (default config file is $HOME/.aws/config)
  -t, --token string      MFA Token for User
  ```
