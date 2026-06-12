# gotak-network — VPC for the ROSA demo (Workspace 1 of 2)

Single-AZ VPC in **us-east-2** that the `gotak-compute` (ROSA Classic) workspace
builds on. This is **Step 1** of the incremental build — it provisions networking
only; no cluster, no platform services.

```
VPC 10.0.0.0/16  (us-east-2a)
├── public  10.0.0.0/24  → IGW        (tag kubernetes.io/role/elb=1)
└── private 10.0.1.0/24  → NAT GW      (tag kubernetes.io/role/internal-elb=1)
```

Outputs consumed by `gotak-compute`: `vpc_id`, `public_subnet_ids`,
`private_subnet_ids`, `availability_zone`, `region`.

## Terraform Cloud setup (one-time)

Workflow is **VCS-driven** — pushes to this repo trigger plans; you confirm applies
in the TFC UI.

1. **Create the workspace** in org `osage`, inside the **goTak**
   project (Projects group workspaces and let you scope a shared Variable Set):
   - Type: **Version control workflow** → connect GitHub repo **`osage-io/gotak`**.
   - **Project:** goTak
   - **Working Directory:** `iac/network`
   - **Branch:** `main`
   - Under *Settings → General*, set **Automatic Run Triggering** so only changes
     under the working directory queue runs (avoids replanning on app commits).
2. **Add AWS credentials once via a project-scoped Variable Set** (so both
   `gotak-network` and `gotak-compute` inherit them — no per-workspace duplication):
   *Org Settings → Variable sets → Create* → scope to the **goTak project** →
   add **Environment** variables, **Sensitive**:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`

   (region defaults to `us-east-2` via the `aws_region` Terraform variable; override
   as a Terraform variable if needed. The compute workspace later adds `RHCS_TOKEN`
   to this same set.)
3. **Enable remote state sharing** so `gotak-compute` can read these outputs:
   *Settings → General → Remote state sharing* → share with the `gotak-compute`
   workspace. Compute will read them via the `tfe_outputs` data source /
   `terraform_remote_state`.

## Apply

```bash
git add iac/network && git commit -m "network: VPC for ROSA demo" && git push
```
TFC queues a plan on the `gotak-network` workspace → review → **Confirm & Apply**.

## Cost note

One NAT gateway (~$0.045/hr + data) and one EIP. Tear down with a **Destroy run**
in TFC when you're done with the demo.

## Overrides

All knobs have defaults (`variables.tf`): `vpc_cidr`, `public_subnet_cidr`,
`private_subnet_cidr`, `availability_zone`, `cluster_name`, `aws_region`. Set any as
Terraform variables in the workspace to change them. `cluster_name` **must match**
the value used in `gotak-compute`.
