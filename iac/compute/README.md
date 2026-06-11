# gotak-compute — ROSA Classic cluster (Workspace 2 of 2)

**Step 2.** Stands up a **ROSA Classic** OpenShift cluster on the VPC built by
`gotak-network`: single-AZ (`us-east-2a`), public endpoint, **3 worker nodes**.
The module also creates the required IAM **account roles, operator roles, and
OIDC** config in the same apply.

Consumes from `gotak-network` (via remote state): `vpc_cidr`, `public_subnet_ids`,
`private_subnet_ids`, `availability_zone`.

> ⚠️ A ROSA Classic apply takes **~30–45 minutes** and provisions **3 control-plane
> + 2 infra + 3 worker** nodes (≈8 × `m5.xlarge`). It's real money — destroy when done.

## Is it ready to go? — Do these first

1. **Create the `gotak-compute` workspace** in the **goTak** project (org
   `org-LoxuyV1DiwAxdXPf`): VCS workflow → repo `osage-io/gotak` →
   **Working Directory `iac/compute`**, branch `main`.

2. **Red Hat OCM token** — the `rhcs` provider needs it:
   - Get an offline token at <https://console.redhat.com/openshift/token>.
   - Add it to the **goTak project Variable Set** as an **Environment** variable,
     **Sensitive**: `RHCS_TOKEN`. (AWS creds are already in that set from step 1.)

3. **Enable ROSA in the AWS account** (one-time, outside Terraform):
   - Accept the ROSA terms / enable the service in the AWS console (ROSA → Get started),
     **or** ensure the ELB service-linked role exists:
     `aws iam create-service-linked-role --aws-service-name elasticloadbalancing.amazonaws.com`
   - Make sure your Red Hat account is linked to this AWS account.

4. **AWS quota** — single-AZ ROSA Classic needs ~**32 vCPU** of On-Demand Standard
   instances in `us-east-2` (8 × m5.xlarge). Bump the *Running On-Demand Standard
   (A, C, D, H, I, M, R, T, Z) instances* quota if you're near the default.

5. **Pick a supported OpenShift version** — set the `openshift_version` Terraform
   variable to a current install version (`rosa list versions`). Default in
   `variables.tf` is a placeholder and may be out of date.

6. **Remote state sharing** — `gotak-network` must share state with this workspace
   (Settings → General → Remote state sharing on `gotak-network`). You enabled this
   in step 1's setup.

7. **`cluster_name` must match** the value used in `gotak-network` (default
   `gotak-demo`) — the subnet role tags are keyed off it.

## Apply

Push `iac/compute/` (PR → merge to `main`) → TFC plans `gotak-compute` →
**Confirm & Apply**. Watch it in the TFC run UI; it will sit "creating" for a while.

## After it's up

- Grab the console with `terraform output console_url`.
- Create a cluster-admin to log in: `rosa create admin -c gotak-demo` (then use the
  printed `oc login` / console credentials).

## Tear down

Run a **Destroy** on `gotak-compute` **first**, then `gotak-network`. (Compute
depends on the network's subnets.)
