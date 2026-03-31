# Consul API Gateway Configuration Entry with TLS
# This defines the API gateway with HTTPS listeners for *.demoland.io

Kind = "api-gateway"
Name = "demoland-gateway"

# Listeners define the ports and protocols the gateway accepts
Listeners = [
  {
    Name     = "https-listener"
    Port     = 8443
    Protocol = "http"
    TLS = {
      Certificates = [
        {
          Kind = "inline-certificate"
          Name = "demoland-wildcard-cert"
        }
      ]
      MinVersion = "TLSv1_2"
      MaxVersion = "TLSv1_3"
      CipherSuites = [
        "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
        "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
        "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
        "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
        "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
        "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"
      ]
    }
  },
  {
    Name     = "cot-tls-listener"
    Port     = 8089
    Protocol = "tcp"
    TLS = {
      Certificates = [
        {
          Kind = "inline-certificate"
          Name = "demoland-wildcard-cert"
        }
      ]
      MinVersion = "TLSv1_2"
    }
  }
]
