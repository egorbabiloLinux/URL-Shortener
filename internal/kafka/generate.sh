#!/usr/bin/env bash
 
set -eu

CN="${CN:-kafka-admin}"
PASSWORD="${PASSWORD:-supersecret}"
TO_GENERATE_PEM="${CITY:-yes}"

VALIDITY_IN_DAYS=3650
CA_WORKING_DIRECTORY="certificate-authority"
TRUSTSTORE_WORKING_DIRECTORY="truststore"
KEYSTORE_WORKING_DIRECTORY="keystore"
PEM_WORKING_DIRECTORY="pem"
CA_KEY_FILE="ca-key"
CA_CERT_FILE="ca-cert"
DEFAULT_TRUSTSTORE_FILE="kafka.truststore.jks"
KEYSTORE_SIGN_REQUEST="cert-file"
KEYSTORE_SIGN_REQUEST_SRL="ca-cert.srl"
KEYSTORE_SIGNED_CERT="cert-signed"
KAFKA_HOSTS_FILE="kafka-hosts.txt"
 
if [ ! -f "$KAFKA_HOSTS_FILE" ]; then
  echo "'$KAFKA_HOSTS_FILE' does not exists. Create this file"
  exit 1
fi
 
if [ -f "$CA_WORKING_DIRECTORY/$CA_KEY_FILE" ] && [ -f "$CA_WORKING_DIRECTORY/$CA_CERT_FILE" ]; then
  echo "Use existing $CA_WORKING_DIRECTORY/$CA_KEY_FILE and $CA_WORKING_DIRECTORY/$CA_CERT_FILE ..."
else
  rm -rf $CA_WORKING_DIRECTORY && mkdir $CA_WORKING_DIRECTORY

  openssl req -new -newkey rsa:4096 -days $VALIDITY_IN_DAYS -x509 -subj "/CN=$CN" \
    -keyout $CA_WORKING_DIRECTORY/$CA_KEY_FILE -out $CA_WORKING_DIRECTORY/$CA_CERT_FILE -nodes
fi

rm -rf $KEYSTORE_WORKING_DIRECTORY && mkdir $KEYSTORE_WORKING_DIRECTORY
while read -r KAFKA_HOST || [ -n "$KAFKA_HOST" ]; do
  if [[ $KAFKA_HOST =~ ^kafka-[0-9]+$ ]]; then
      SUFFIX="server"
      DNAME="CN=$KAFKA_HOST"
  else
      SUFFIX="client"
      DNAME="CN=client"
  fi
  KEY_STORE_FILE_NAME="$KAFKA_HOST.$SUFFIX.keystore.jks"

  keytool -genkey -keystore $KEYSTORE_WORKING_DIRECTORY/"$KEY_STORE_FILE_NAME" \
    -alias localhost -validity $VALIDITY_IN_DAYS -keyalg RSA \
    -noprompt -dname $DNAME -keypass $PASSWORD -storepass $PASSWORD

  keytool -certreq -keystore $KEYSTORE_WORKING_DIRECTORY/"$KEY_STORE_FILE_NAME" \
    -alias localhost -file $KEYSTORE_SIGN_REQUEST -keypass $PASSWORD -storepass $PASSWORD
 
  openssl x509 -req -CA $CA_WORKING_DIRECTORY/$CA_CERT_FILE \
    -CAkey $CA_WORKING_DIRECTORY/$CA_KEY_FILE \
    -in $KEYSTORE_SIGN_REQUEST -out $KEYSTORE_SIGNED_CERT \
    -days $VALIDITY_IN_DAYS -CAcreateserial
 
  keytool -keystore $KEYSTORE_WORKING_DIRECTORY/"$KEY_STORE_FILE_NAME" -alias CARoot \
    -import -file $CA_WORKING_DIRECTORY/$CA_CERT_FILE -keypass $PASSWORD -storepass $PASSWORD -noprompt
 
  keytool -keystore $KEYSTORE_WORKING_DIRECTORY/"$KEY_STORE_FILE_NAME" -alias localhost \
    -import -file $KEYSTORE_SIGNED_CERT -keypass $PASSWORD -storepass $PASSWORD

  rm -f $CA_WORKING_DIRECTORY/$KEYSTORE_SIGN_REQUEST_SRL $KEYSTORE_SIGN_REQUEST $KEYSTORE_SIGNED_CERT
done < "$KAFKA_HOSTS_FILE"

rm -rf $TRUSTSTORE_WORKING_DIRECTORY && mkdir $TRUSTSTORE_WORKING_DIRECTORY
keytool -keystore $TRUSTSTORE_WORKING_DIRECTORY/$DEFAULT_TRUSTSTORE_FILE \
  -alias CARoot -import -file $CA_WORKING_DIRECTORY/$CA_CERT_FILE \
  -noprompt -dname "CN=$CN" -keypass $PASSWORD -storepass $PASSWORD

if [ "$TO_GENERATE_PEM" == "yes" ]; then
  rm -rf "$PEM_WORKING_DIRECTORY" && mkdir "$PEM_WORKING_DIRECTORY"

  # Найти первый .client.keystore.jks файл
  CLIENT_KEYSTORE=$(find "$KEYSTORE_WORKING_DIRECTORY" -name "*.client.keystore.jks" | head -n 1)

  if [ -z "$CLIENT_KEYSTORE" ]; then
    echo "Не найден .client.keystore.jks для генерации PEM-файлов."
    exit 1
  fi

  echo "Используется клиентский keystore: $CLIENT_KEYSTORE"

  keytool -exportcert -alias CARoot -keystore "$CLIENT_KEYSTORE" \
    -rfc -file "$PEM_WORKING_DIRECTORY/ca-root.pem" -storepass "$PASSWORD"

  keytool -exportcert -alias localhost -keystore "$CLIENT_KEYSTORE" \
    -rfc -file "$PEM_WORKING_DIRECTORY/client-certificate.pem" -storepass "$PASSWORD"

  keytool -importkeystore -srcalias localhost -srckeystore "$CLIENT_KEYSTORE" \
    -destkeystore cert_and_key.p12 -deststoretype PKCS12 -srcstorepass "$PASSWORD" -deststorepass "$PASSWORD"

  openssl pkcs12 -in cert_and_key.p12 -nocerts -nodes -password pass:$PASSWORD \
    | awk '/-----BEGIN PRIVATE KEY-----/,/-----END PRIVATE KEY-----/' > "$PEM_WORKING_DIRECTORY/client-private-key.pem"

  rm -f cert_and_key.p12
fi
