# GoSpray
Active Directory Password Spray Testing Utility in Go

## Install

```
git clone https://github.com/wunderwuzzi23/gospray
```

GoSpray has a dependency on ldap and gomail (via wuzziutils library). You can get those dependencies with:
```
go get github.com/wunderwuzzi23/wuzziutils
go get gopkg.in/ldap.v3
go get gopkg.in/gomail.v2
```
## Building

Navigate to the `gospray` folder and run:

```
go build gospray.go
```

## Running

./gospray [options]

There are various command line options to specify user and password list, as well as LDAP server, certificates (when testing from non domain joined machines). Nothing fancy.

At a high level the latest version supports two testing modes:
1. **Password Spray:** If you specify both *-accounts* and *-passwords* files, then a spray will be performed
2. **Password Validation Mode:** If you specify *-validatecreds* file, the above options are ignored. The file specified with validatecreds   is parsed line by line, each line is split by colon (:) to retrieve username:password. Afterwards an authentication attempt will be performed against specified domain controller.


## Testing from a non domain joined machine? You will need the domain controller's TLS cert

When on non-domain joined machine, it's possible to get the domain controllers cert via (basically trust on first use - tofu):
```
openssl s_client -showcerts -connect [dc].[company].[local]:636 > cer.txt
```
Afterwards the pem part of the result in a `ca.pem` file to gospray and it will use it to valide the servers certificate.

Otherwise ask your IT deparment or copy cert from your domain joined machine.

## Note

**Credential validation can cause account lockouts. Please use with care, have authorizatoin and know what you do.**

## Email Notification

To have GoSpray send emails, manually update the SMTP information in the `gospray.go` file. I haven't had the time to make that configurable via command line option


Have fun!
