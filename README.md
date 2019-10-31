# awsmfa

Small commandline tool based on spf13/Cobra cli library to generate AWS Session Credentials for IAM Roles and IAM Users that require MFA.

It will by default save credentials in Users Home directory. 

## Usage
```
Cobra cli based app to create aws session credentials , currently supports IAM roles
configured in default (~/.aws/config)
Also IAM User Sessiont Credentials with virtual mfa

Usage:
  awsmfa [flags]
  awsmfa [command]

Available Commands:
  help        Help about any command
  roleSession Manage your AWS Session Credentials for IAM Roles with MFA enabled
  userSession aws sts get-session-token implementation

Flags:
  -h, --help   help for awsmfa

Use "awsmfa [command] --help" for more information about a command.

  ```
