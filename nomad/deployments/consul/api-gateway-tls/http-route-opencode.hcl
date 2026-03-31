# HTTP Route Configuration Entry for opencode.demoland.io
# Routes HTTPS traffic from the API gateway to opencode service

Kind = "http-route"
Name = "opencode-demoland-route"

# Bind to the API gateway's HTTPS listener
Parents = [
  {
    Kind        = "api-gateway"
    Name        = "demoland-gateway"
    SectionName = "https-listener"
  }
]

# Route based on hostname
Hostnames = ["opencode.demoland.io"]

# Route all requests to opencode service
Rules = [
  {
    Matches = [
      {
        Path = {
          Match = "prefix"
          Value = "/"
        }
      }
    ]
    Services = [
      {
        Name   = "opencode"
        Weight = 100
      }
    ]
  }
]
