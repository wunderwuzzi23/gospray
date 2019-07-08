# GoSpray
Active Directory Password Spray Testing Utility in Go

This tool has a dependency on ldap, please install the following library:
*go get gopkg.in/ldap.v3*

The first  version is loading users.list and password.list and trying the variations against the LDAP server provided in code. Nothing fancy.

**Note:** Credential validation can cause account lockouts. Please use with care, have authorizatoin and know what you do.

Have fun!
