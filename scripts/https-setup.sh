#!/usr/bin/env sh

set -eu

CONFIG_FILE="${CONFIG_FILE:-_datafiles/config.yaml}"

if [ ! -f "$CONFIG_FILE" ]; then
	echo "Config file not found: $CONFIG_FILE" >&2
	exit 1
fi

current_web_domain=$(awk -F': ' '/^[[:space:]]*WebDomain:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$CONFIG_FILE")
current_https_email=$(awk -F': ' '/^[[:space:]]*HttpsEmail:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$CONFIG_FILE")
current_http_port=$(awk -F': ' '/^[[:space:]]*HttpPort:/ {print $2; exit}' "$CONFIG_FILE")
current_https_port=$(awk -F': ' '/^[[:space:]]*HttpsPort:/ {print $2; exit}' "$CONFIG_FILE")
current_https_redirect=$(awk -F': ' '/^[[:space:]]*HttpsRedirect:/ {print $2; exit}' "$CONFIG_FILE")
current_https_cert=$(awk -F': ' '/^[[:space:]]*HttpsCertFile:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$CONFIG_FILE")
current_https_key=$(awk -F': ' '/^[[:space:]]*HttpsKeyFile:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$CONFIG_FILE")
current_https_cache_dir=$(awk -F': ' '/^[[:space:]]*HttpsCacheDir:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$CONFIG_FILE")

printf 'Interactive HTTPS setup\n'
printf 'Config file: %s\n\n' "$CONFIG_FILE"

printf 'Choose HTTPS mode:\n'
printf '  1) Automatic Let'"'"'s Encrypt (recommended)\n'
printf '  2) Manual certificate files\n'
printf '  3) HTTP only\n'
printf 'Selection [1]: '
IFS= read -r mode_selection
mode_selection=${mode_selection:-1}

web_domain=$current_web_domain
https_email=$current_https_email
https_cert_file=$current_https_cert
https_key_file=$current_https_key
https_cache_dir=${current_https_cache_dir:-_datafiles/tls}
http_port=$current_http_port
https_port=$current_https_port
https_redirect=$current_https_redirect

case "$mode_selection" in
1)
	printf 'Target domain name [%s]: ' "${current_web_domain:-play.example.com}"
	IFS= read -r web_domain_input
	if [ -n "$web_domain_input" ]; then
		web_domain=$web_domain_input
	elif [ -z "$web_domain" ]; then
		web_domain=play.example.com
	fi

	printf 'Contact email for Let'"'"'s Encrypt notices [%s]: ' "${current_https_email:-admin@example.com}"
	IFS= read -r https_email_input
	if [ -n "$https_email_input" ]; then
		https_email=$https_email_input
	elif [ -z "$https_email" ]; then
		https_email=admin@example.com
	fi

	printf 'Certificate cache directory [%s]: ' "${https_cache_dir:-_datafiles/tls}"
	IFS= read -r https_cache_input
	if [ -n "$https_cache_input" ]; then
		https_cache_dir=$https_cache_input
	fi

	printf 'Redirect HTTP to HTTPS? [%s]: ' "${current_https_redirect:-true}"
	IFS= read -r redirect_input
	if [ -n "$redirect_input" ]; then
		https_redirect=$redirect_input
	else
		https_redirect=${current_https_redirect:-true}
	fi

	http_port=80
	https_port=443
	https_cert_file=""
	https_key_file=""
	;;
2)
	printf 'Public domain name [%s]: ' "${current_web_domain:-play.example.com}"
	IFS= read -r web_domain_input
	if [ -n "$web_domain_input" ]; then
		web_domain=$web_domain_input
	elif [ -z "$web_domain" ]; then
		web_domain=play.example.com
	fi

	printf 'Certificate file path [%s]: ' "${current_https_cert:-server.crt}"
	IFS= read -r cert_input
	if [ -n "$cert_input" ]; then
		https_cert_file=$cert_input
	elif [ -z "$https_cert_file" ]; then
		https_cert_file=server.crt
	fi

	printf 'Private key file path [%s]: ' "${current_https_key:-server.key}"
	IFS= read -r key_input
	if [ -n "$key_input" ]; then
		https_key_file=$key_input
	elif [ -z "$https_key_file" ]; then
		https_key_file=server.key
	fi

	printf 'Redirect HTTP to HTTPS? [%s]: ' "${current_https_redirect:-true}"
	IFS= read -r redirect_input
	if [ -n "$redirect_input" ]; then
		https_redirect=$redirect_input
	else
		https_redirect=${current_https_redirect:-true}
	fi

	printf 'HTTP port [%s]: ' "${current_http_port:-80}"
	IFS= read -r http_port_input
	if [ -n "$http_port_input" ]; then
		http_port=$http_port_input
	elif [ -z "$http_port" ]; then
		http_port=80
	fi

	printf 'HTTPS port [%s]: ' "${current_https_port:-443}"
	IFS= read -r https_port_input
	if [ -n "$https_port_input" ]; then
		https_port=$https_port_input
	elif [ -z "$https_port" ]; then
		https_port=443
	fi

	https_email=""
	;;
3)
	printf 'HTTP port [%s]: ' "${current_http_port:-80}"
	IFS= read -r http_port_input
	if [ -n "$http_port_input" ]; then
		http_port=$http_port_input
	elif [ -z "$http_port" ]; then
		http_port=80
	fi

	https_port=0
	https_redirect=false
	https_email=""
	https_cert_file=""
	https_key_file=""
	;;
*)
	echo "Unknown selection: $mode_selection" >&2
	exit 1
	;;
esac

printf '\nPlanned settings:\n'
printf '  WebDomain: %s\n' "$web_domain"
printf '  HttpsEmail: %s\n' "$https_email"
printf '  HttpsCertFile: %s\n' "${https_cert_file:-<empty>}"
printf '  HttpsKeyFile: %s\n' "${https_key_file:-<empty>}"
printf '  HttpsCacheDir: %s\n' "$https_cache_dir"
printf '  HttpPort: %s\n' "$http_port"
printf '  HttpsPort: %s\n' "$https_port"
printf '  HttpsRedirect: %s\n' "$https_redirect"

printf '\nApply these changes to %s? [Y/n]: ' "$CONFIG_FILE"
IFS= read -r confirm
confirm=${confirm:-Y}
case "$confirm" in
Y | y | yes | YES)
	;;
*)
	echo "Aborted without changing the config."
	exit 0
	;;
esac

timestamp=$(date +%Y%m%d%H%M%S)
backup_file="${CONFIG_FILE}.bak.${timestamp}"
cp "$CONFIG_FILE" "$backup_file"

tmp_file=$(mktemp "${TMPDIR:-/tmp}/gomud-https-setup.XXXXXX")

awk \
	-v web_domain="$web_domain" \
	-v https_email="$https_email" \
	-v https_cert_file="$https_cert_file" \
	-v https_key_file="$https_key_file" \
	-v https_cache_dir="$https_cache_dir" \
	-v http_port="$http_port" \
	-v https_port="$https_port" \
	-v https_redirect="$https_redirect" \
	'
  function yaml_string(s, escaped) {
    escaped = s
    gsub(/\\/,"\\\\", escaped)
    gsub(/"/,"\\\"", escaped)
    return "\"" escaped "\""
  }
  /^[[:space:]]*WebDomain:/      { print "  WebDomain: " yaml_string(web_domain); next }
  /^[[:space:]]*HttpsEmail:/     { print "  HttpsEmail: " yaml_string(https_email); next }
  /^[[:space:]]*HttpsCertFile:/  { print "  HttpsCertFile: " yaml_string(https_cert_file); next }
  /^[[:space:]]*HttpsKeyFile:/   { print "  HttpsKeyFile: " yaml_string(https_key_file); next }
  /^[[:space:]]*HttpsCacheDir:/  { print "  HttpsCacheDir: " yaml_string(https_cache_dir); next }
  /^[[:space:]]*HttpPort:/       { print "  HttpPort: " http_port; next }
  /^[[:space:]]*HttpsPort:/      { print "  HttpsPort: " https_port; next }
  /^[[:space:]]*HttpsRedirect:/  { print "  HttpsRedirect: " https_redirect; next }
  { print }
  ' "$CONFIG_FILE" >"$tmp_file"

mv "$tmp_file" "$CONFIG_FILE"
mkdir -p "$https_cache_dir"

printf '\nHTTPS setup updated.\n'
printf 'Backup saved to: %s\n' "$backup_file"
printf 'Next steps:\n'
case "$mode_selection" in
1)
	printf '  1. Point DNS for %s at this server.\n' "$web_domain"
	printf '  2. Open inbound TCP ports 80 and 443.\n'
	printf '  3. Start GoMud and visit /admin/https/ to confirm the certificate is issued.\n'
	;;
2)
	printf '  1. Make sure %s and %s exist and are readable.\n' "$https_cert_file" "$https_key_file"
	printf '  2. Start GoMud and visit /admin/https/ to confirm manual HTTPS is active.\n'
	;;
3)
	printf '  1. Start GoMud and connect over plain HTTP on port %s.\n' "$http_port"
	;;
esac
