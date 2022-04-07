
mmdb:
	wget -O GeoLite2-City.mmdb.tar.gz \
		"https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$(LICENSE_KEY)&suffix=tar.gz"

deploy:
	fly deploy

