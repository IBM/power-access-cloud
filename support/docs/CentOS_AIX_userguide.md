# CentOS & AIX User Guide

Welcome to the CentOS & AIX User Guide for IBM&reg; Power&reg; Access Cloud. This guide walks you through deploying CentOS or AIX instances and accessing them via SSH.

---

## 1. Phase 1: Deployment via Power Access Cloud

**Steps:**

### Access Catalog
1. Log in to your Power Access Cloud portal.
2. Navigate to the **Catalog** by clicking on the "Catalog" button or "Go to Catalog".

### Select Image
1. Locate the CentOS or AIX base image (or the specific version required for your workload).

### Configure & Deploy
1. Click on the **Deploy** button.
2. Enter the name for your Virtual Machine and click **Submit** to start provisioning.

### Monitor Status
1. Go to the **Services** tab on the Home page.
2. The status will initially show as **Deploying**.
3. Wait for the status to transition to **Active**.

---

## 2. Phase 2: Boot Process

**Important Note:**

Once the status shows **Active**, an External IP address will be assigned to your machine.

---

## 3. Phase 3: SSH Access

**Steps:**

Once the status is active and the External IP is assigned, you can connect to your VM via SSH.

1. **Retrieve IP:** Copy the External IP from the service details page.
2. **Open Local Terminal:** Open your local terminal (PowerShell, Command Prompt, or Terminal on your workstation).
3. **Connect via SSH:** Run the following command:
   ```bash
   ssh root@<external-ip>
   ```
   Replace `<external-ip>` with the actual External IP address of your VM.

**Example:**
```bash
ssh root@150.240.64.10
```

4. **Accept Host Key:** On first connection, you'll be prompted to accept the host key. Type `yes` and press Enter.
5. **You're In:** You should now be logged into your CentOS or AIX VM.

**Troubleshooting:**

If you encounter a "Host key verification failed" error, run the following command to remove the old host key from the known hosts file:

```bash
ssh-keygen -R <your-external-ip>
```

---

## Notes

- The default user for SSH access is `root`.
- Ensure your SSH client is properly configured on your local machine.
- If you encounter connection issues, wait a few more minutes for the VM to fully boot and the SSH service to start.

---