# GoSpray
Active Directory Password Spray Testing Utility in Go

This tool has a dependency on ldap and gomail (via wuzziutils library), please install the following library:

*go get gopkg.in/ldap.v3*

*go get gopkg.in/gomail.v2*


There are various command line options to specify user and password list, as well as LDAP server, certificates (when testing from non domain joined machines). Nothing fancy.

At a high level the latest version supports two testing modes:
1. Password Spray: If you specify both -accounts and -passwords files, then a spray will be performed
2. Password Validation Mode: If you specify -validatecreds file (the above options are ignored), and the validatecreds file will be parsed line by line, each line is split by colon (:), retrieving username:password, and afterwards and authentication attempt will be performed against specified DC

When on non-domain joined machine, it's possible to get the domain controllers cert via (basically trust on first use - tofu):

openssl s_client -showcerts -connect [dc].[company].[local]:636 > cer.txt

Afterwards the pem part of the result in a ca.pem file to gospray and it will use it to valide the servers certificate.

Otherwise ask your IT deparment or copy cert from your domain joined machine.

**Note:** Credential validation can cause account lockouts. Please use with care, have authorizatoin and know what you do.

Have fun!
