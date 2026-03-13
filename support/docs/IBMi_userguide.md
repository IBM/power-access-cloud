# IBM&reg; i User Guide

Welcome to the IBM&reg; i User Guide for IBM&reg; Power&reg; Access Cloud. This guide walks you through deploying an IBM&reg; i instance and accessing it via SSH and 5250 console.

---
A guided video instruction:

[Watch the guided video instruction](https://github.com/user-attachments/assets/94383e64-b948-457f-a761-d3a1660f8e7b)

## 1. Phase 1: Deployment via Power Access Cloud

**Steps:**

### Access Catalog
1. Log in to your Power Access Cloud portal.
2. Navigate to the **Catalog** by clicking on the "Catalog" button or "Go to Catalog".

### Select Image
1. Locate the IBM i base image (or the specific version required for your workload).

### Configure & Deploy
1. Click on the **Deploy** button.
2. Enter the name for your Virtual Machine and click **Submit** to start provisioning.

### Monitor Status
1. Go to the **Services** tab on the Home page.
2. The status will initially show as **Deploying**.
3. Wait for the status to transition to **Active**.

---

## 2. Phase 2: The IPL and Boot Process

**Important Note:**

Once the status shows **Active**, an External IP address will be assigned to your machine.

**For IBM i-based VMs**, the system must undergo an Initial Program Load (IPL) and several internal boot-up sequences. It typically takes **30–40 minutes** for the status to turn "Active" before the SSHD (SSH Daemon) starts and the machine becomes accessible via the network.

---

## 3. Phase 3: Initial Remote Access

**Steps:**

Once the status is active, you can establish your first connection using the default administrative profile.

1. **Retrieve IP:** Copy the External IP from the service details page.
2. **Connect via SSH:** Open your local terminal (PowerShell, Command Prompt, or Terminal) and run:
   ```bash
   ssh PAC@<your-external-ip>
   ```

---

## 4. Accessing 5250 Console via SSH-Key Tunneling

**Prerequisites:** You have successfully enabled SSH-key based login.

[Watch the guided video instruction](https://github.com/user-attachments/assets/1d3f86f0-8b48-443d-8249-208af8875550)

Follow these steps to establish your 5250 console:

### Step 1: Set IBM i User Password

SSH keys handle the secure "handshake" for the tunnel, but the 5250 sign-on screen still requires a standard IBM i password.

```
system "CHGUSRPRF USRPRF(PAC) PASSWORD(YourSecurePassword)"
```

### Step 2: Enable Telnet Service

The 5250 session runs over Telnet (Port 23), which must be active on the system to accept the tunneled connection.

```
system "STRTCPSVR *TELNET"
```

### Step 3: Establish the SSH Tunnel

On your local workstation (Laptop/PC), open a terminal or command prompt to create the secure bridge. This command maps a local port to the IBM i Telnet port.

```bash
ssh -L 50000:localhost:23 -L 2001:localhost:2001  -L 449:localhost:449 -L 8470:localhost:8470 -L 8471:localhost:8471 -L 8472:localhost:8472 -L 2007:localhost:2007 -L 8473:localhost:8473 -L 8474:localhost:8474 -L 8475:localhost:8475 -L 8476:localhost:8476 -L 2003:localhost:2003 -L 2002:localhost:2002 -L 2006:localhost:2006 -L 2300:localhost:2300 -L 2323:localhost:2323 -L 3001:localhost:3001 -L 3002:localhost:3002 -L 2005:localhost:2005  -o ExitOnForwardFailure=yes -o ServerAliveInterval=15 -o ServerAliveCountMax=3 PAC@powervs_public_ip
```
**Note:** For furture reference visit https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-connect-ibmi#ssh-tunneling and also some ports has restricted access so recommended to use sudo.

**Parameters:**
- `-L`: Maps local port 50000 to the remote localhost port 23.

**Important:** Keep this terminal window open; closing it kills the connection to the 5250 console.

### Step 4: Configure ACS 5250 Session

With the tunnel active, configure your IBM i Access Client Solutions (ACS) tool to point at the local end of the tunnel.

1. Open the **ACS Tool**.
2. Go to the **5250 Session Manager**.
3. Select **New Display Session**.
4. In the settings, enter the following:
   - **Destination:** `localhost`
   - **Port:** `50000` (This must match the port used in your SSH tunnel command).
5. Click **OK** or **Connect**.

### Step 5: Sign On to 5250

The classic IBM i "Sign On" display will appear.

1. Enter your **User ID** (PAC).
2. Enter the **Password** you created in Step 1.