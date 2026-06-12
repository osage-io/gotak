# Extra Nomad client config merged into the dev agent (see up.sh).
#
# On Apple Silicon, Nomad's CPU fingerprinter reports an absurdly low
# cpu.totalcompute (~40 MHz), which starves real jobs of CPU and causes
# "Dimension cpu exhausted" placement failures. Pin a sane total compute
# so jobs that request hundreds of MHz can actually be placed.
client {
  cpu_total_compute = 40000
}
