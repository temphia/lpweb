# lpweb

lpweb is a way to access http service though libp2p. 
This allows to expose your http service even if you only have private ip like serices running in your local machine or your home rapPi. 
It gives you domain name using your libp2p public key.

<libp2p_key_hash>.lpweb


browser => ( proxy_or_extension_with_libp2p) => (another_libp2p_node_with_http_service)
