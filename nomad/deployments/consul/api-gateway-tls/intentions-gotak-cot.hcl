# Service Intentions for gotak-cot service
# Allows the gateway to route CoT traffic

Kind = "service-intentions"
Name = "gotak-cot"
Sources = [
  {
    Name   = "demoland-gateway"
    Action = "allow"
  }
]
