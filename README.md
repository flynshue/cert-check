# eck-cert-check
POC

# Using script to generate certs and keystore
```bash
./gen-keystore.sh

```

After running the script you should see a new `cert/` directory
```bash
certs
├── ca-cert.crt
├── ca.key
├── new
│   ├── consumer-sa.crt
│   ├── consumer-sa.csr
│   ├── consumer-sa.key
│   ├── keystore.jks
│   ├── logger-sa.crt
│   └── logger-sa.csr
├── old
│   ├── consumer-sa.crt
│   ├── consumer-sa.csr
│   ├── consumer-sa.key
│   ├── keystore.jks
│   ├── logger-sa.crt
│   └── logger-sa.csr
└── storepass
```

# Creating java keystore certs
Create key/pair
```bash
$ keytool -genkey -keystore keystore.jks -alias localhost -keyalg RSA -validity 14 -storetype JKS
Enter keystore password:  
Re-enter new password: 
What is your first and last name?
  [Unknown]:  logging-sa
What is the name of your organizational unit?
  [Unknown]:  Infrastructure
What is the name of your organization?
  [Unknown]:  Fake Org
What is the name of your City or Locality?
  [Unknown]:  Brooms Town
What is the name of your State or Province?
  [Unknown]:  
What is the two-letter country code for this unit?
  [Unknown]:  US
Is CN=logging-sa, OU=Infrastructure, O=Fake Org, L=Brooms Town, ST=Unknown, C=US correct?
  [no]:  yes

Enter key password for <localhost>
        (RETURN if same as keystore password):  

Warning:
The JKS keystore uses a proprietary format. It is recommended to migrate to PKCS12 which is an industry standard format using "keytool -importkeystore -srckeystore keystore.jks -destkeystore keystore.jks -deststoretype pkcs12".
```

Verify contents in keystore
```bash
keytool -list -v -keystore keystore.jks
```

Generate Root CA cert
```bash
$ openssl req -new -x509 -days 3650 -keyout ca.key -out ca-cert.crt
........+......+.+.....+.........+.+......+...+..+...+....+.....+.......+.....+....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*......+......+....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.........+.....+......+.......+...+..+.......+..+.+.........+...+...............+...+............+...+...+.....+.+.....+.+..+.+.................+....+..............+......+.+..............+..........+..................+..+.+..+............+.+...+..+.......+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
.+...+....+..+................+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.....+.+.....+.+...+........+...+....+..+....+..............+...+...+......+.+..+.......+...+...+.....+.............+..+...............+...+.+...+......+......+...+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+..+.+.....+......+.+..+..........+.....+..........+......+.....+.......+...+..+..................+..........+.........+..+.+..+.......+.....+....+.................+.......+......+.........+..+...+...+....+........+......+.+........+...+...+.........+.+......+.........+..+....+.....+.............+..+.............+..+.+..............+...+..........+...+..+......+......+.........+.........+.+.....+....+...+..+......+.+...+........+.............+.........+......+.....+.......+..+.+..+...+.......+..+.......+......+.....+..........+..+............+....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
Enter PEM pass phrase:
Verifying - Enter PEM pass phrase:
-----
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:US
State or Province Name (full name) [Some-State]:
Locality Name (eg, city) []:Brooms Town
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Fake Org
Organizational Unit Name (eg, section) []:Infrastructure
Common Name (e.g. server FQDN or YOUR name) []:fake-kafka-01a.fakeorg.us
Email Address []:fake-infra@fakeorg.us
```

Generate cert request from keytool so it can be signed be CA
```bash
keytool -keystore keystore.jks -certreq -alias localhost -file cert-file
```

Signing the cert request from keytool
```bash
openssl x509 -req -CA ca-cert.crt -CAkey ca.key -in cert-file -out cert-signed.crt -days 14 -CAcreateserial -passin pass:foobar
```

Import the CA and client cert key pair into keytool
```bash
keytool -importcert -alias caroot -file ca-cert.crt -keystore old/keystore.jks

keytool -importcert -alias localhost -file old/cert-signed.crt -keystore old/keystore.jks
```

Generate CSR for testing with secondary cert outside of keystore
```bash
openssl req -new \
-newkey rsa:2048 -nodes -keyout old/consumer-sa.key \
-out old/consumer-sa.csr \
-subj "/C=US/O=Fake Org/OU=Infrastructure/CN=consumer-sa"
```

Use CA to sign cert
```bash
openssl x509 -req -CA ca-cert.crt -CAkey ca.key -in old/consumer-sa.csr -out old/consumer-sa-cert-signed.csr -days 14 -CAcreateserial -passin pass:foobar
```

# References
* https://docs.confluent.io/platform/current/security/security_tutorial.html#generating-keys-certs
*