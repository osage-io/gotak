# Service Intentions for opencode service
# Allows the gateway to route traffic to opencode

Kind = "service-intentions"
Name = "opencode"
Sources = [
  {
    Name   = "demoland-gateway"
    Action = "allow"
  }
]
