# TCP Route Configuration Entry for CoT TLS
# Routes TLS-encrypted CoT traffic to gotak-cot service

Kind = "tcp-route"
Name = "gotak-cot-tls-route"

# Bind to the API gateway's CoT TLS listener
Parents = [
  {
    Kind        = "api-gateway"
    Name        = "demoland-gateway"
    SectionName = "cot-tls-listener"
  }
]

# Route to gotak-cot service
Services = [
  {
    Name = "gotak-cot"
  }
]
