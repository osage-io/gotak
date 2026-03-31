# TCP Route Configuration Entry
# Routes TCP traffic from the API gateway to gotak-cot service

Kind = "tcp-route"
Name = "gotak-cot-route"

# Bind to the API gateway's TCP listener
Parents = [
  {
    Kind        = "api-gateway"
    Name        = "gotak-gateway"
    SectionName = "cot-listener"
  }
]

# Route to the gotak-cot service
Services = [
  {
    Name = "gotak-cot"
  }
]
