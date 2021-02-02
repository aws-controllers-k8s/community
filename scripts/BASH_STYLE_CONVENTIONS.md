Bash Style Conventions
======================

This guide is meant to outline some conventions when writing bash scripts for this folder. 
The goal is to encourage everyone to keep a minimum of consistency between scripting styles.

**Please participate**. The guide is open for adding new good practices and thoughts.

Conventions
-----------

### Script Arguments

We prefer to have a fixed number of arguments and rely on environment variables for optional behaviour switches instead of using alternatives like `getops`. An example would be the helm-package-controller.sh script:

``` bash
 USAGE=" 
 Usage: 
   $(basename "$0") <service> 
  
 <service> should be an AWS service API aliases that you wish to build -- e.g. 
 's3' 'sns' or 'sqs' 
  
 Environment variables: 
   CHART_INPUT_PATH:         Specify a path for the Helm chart to use as input. 
                             Default: services/{SERVICE}/helm 
   PACKAGE_OUTPUT_PATH:      Specify a path for the Helm chart package to output to. 
                             Default: $BUILD_DIR/release/{SERVICE} 
 " 
  
 if [ $# -ne 1 ]; then 
     echo "ERROR: $(basename "$0") accepts one parameter, the SERVICE" 1>&2 
     echo "$USAGE" 
     exit 1 
 fi 
```

### How the `SCRIPTS|THIS_DIR` and `ROOT_DIR` variables are determined

Use:

``` bash
THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/.."
```

It is consistent with how it's done in other folders like `/test`.
It works even if the scripts are sourced. Otherwise, with alternatives that make use of `dirname $0`, it will point to the bash binary when sourced instead, e.g.:

``` bash
#!/usr/bin/env bash
# test.sh 

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo $THIS_DIR
SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
echo $SCRIPTS_DIR
```

Then, 

```
[user@localhost ]$ ./test.sh 
/home/user/aws-controllers-k8s
/home/user/aws-controllers-k8s
[user@localhost aws-controllers-k8s]$ . ./test.sh 
/home/user/aws-controllers-k8s
/bin
```

### Default values for variables


Use `VAR=${VAR:-default}`:

``` bash
AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID:-""}
```

Makes it easier for reading and Bash only will test for a unset or null parameter before substitute it.


### Error handling

Use:

``` bash
set -eo pipefail
```

Avoid explicitly set `-x`, it will make the default output of the scripts extremely verbose. `-u` should also be skipped, there are a number of scripts that assume the presence of variables (in a "safe" manner). The -u option will just print a whole bunch of error messages that really aren't errors.
