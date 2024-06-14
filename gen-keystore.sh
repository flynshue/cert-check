#!/bin/bash

CERT_DIR=certs
OLD_PATH=${CERT_DIR}/old
NEW_PATH=${CERT_DIR}/new
KEYSTORE_NAME=keystore.jks
STOREPASS=foobar
OU=Infra
ORG="Fake ORG"
COUNTRY=US
LOGGER=logger-sa
CONSUMER=consumer-sa

create_keystore() {
  keystore_file=${1}
  days=${2}
  dname=${3}
  echo "create_keystore"
  keytool -genkey -keystore ${keystore_file} -alias localhost -keyalg RSA \
  -validity ${days} -storetype JKS -storepass ${STOREPASS} \
  -dname "cn=${dname}, ou=${OU}, o=${ORG}, c=${COUNTRY}" -keypass ${STOREPASS}

  echo -n ${STOREPASS} > ${CERT_DIR}/storepass
}

create_rootca() {
  days=${1}
  cakey=${2}
  cacert=${3}
  echo "create_rootca"
  openssl req -new -x509 -days ${days} -keyout ${cakey} -out ${cacert} -passout pass:${STOREPASS} \
  -subj "/C=${COUNTRY}/O=${ORG}/OU=${OU}/CN=fake-kafka-01a.fakeorg.us"
}

create_keytool_req() {
  keystore_file=${1}
  out_file=${2}
  echo "create_keytool_req"
  keytool -keystore ${keystore_file} -certreq -alias localhost -file ${out_file} -storepass ${STOREPASS}
}

sign_cert() {
  csr=${1}
  out_file=${2}
  days=${3}
  cakey=${4}
  cacert=${5}
  echo "sign_cert"
  openssl x509 -req -CA ${cacert} -CAkey ${cakey} -in ${csr} -out ${out_file} -days ${days} -CAcreateserial -passin pass:${STOREPASS}
}

import_cert() {
  alias=${1}
  cert_file=${2}
  keystore=${3}
  echo "import_cert"
  keytool -importcert -alias ${alias} -file ${cert_file} -noprompt \
  -keystore ${keystore} -storepass ${STOREPASS}
}

gen_csr() {
  dname=${1}
  out_key=${2}
  out_csr=${3}
  echo "gen_csr"
  openssl req -new \
  -newkey rsa:2048 -nodes -keyout ${out_key} \
  -out ${out_csr} \
  -subj "/C=US/O=Fake Org/OU=Infrastructure/CN=${dname}"
}

if [ ! -d ${OLD_PATH} ]; then
  mkdir -p ${OLD_PATH}
fi

if [ ! -d ${NEW_PATH} ]; then
  mkdir -p ${NEW_PATH}
fi

generate_cert_package() {
  certpath=${1}
  days=${2}
  create_rootca ${days} ${certpath}/ca.key ${certpath}/ca-cert.crt
  create_keystore ${certpath}/${KEYSTORE_NAME} ${days} ${LOGGER}

  create_keytool_req ${certpath}/${KEYSTORE_NAME} ${certpath}/${LOGGER}.csr

  sign_cert ${certpath}/${LOGGER}.csr ${certpath}/${LOGGER}.crt ${days} ${certpath}/ca.key ${certpath}/ca-cert.crt

  import_cert caroot  ${certpath}/ca-cert.crt ${certpath}/${KEYSTORE_NAME}

  import_cert localhost ${certpath}/${LOGGER}.crt ${certpath}/${KEYSTORE_NAME}

  gen_csr ${CONSUMER} ${certpath}/${CONSUMER}.key ${certpath}/${CONSUMER}.csr

  sign_cert ${certpath}/${CONSUMER}.csr ${certpath}/${CONSUMER}.crt ${days} ${certpath}/ca.key ${certpath}/ca-cert.crt
}

generate_cert_package ${OLD_PATH} 14

generate_cert_package ${NEW_PATH} 60
