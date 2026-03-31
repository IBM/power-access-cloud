# IBM i User Guide

Welcome to the IBM i User Guide for IBM&reg; Power&reg; Access Cloud. This guide walks you through deploying an IBM i instance and accessing it via SSH and 5250 console.

---
A guided video instruction:

[Watch the guided video instruction](https://github.com/user-attachments/assets/94383e64-b948-457f-a761-d3a1660f8e7b)

---

## Pre-Configured Features

Your IBM i machine from Power Access Cloud comes with the following pre-configured features and settings:

### System Information
- **IBM i Version:** 7.5
- **Language:** 2924

### Network Services
The following network services are **auto-enabled** by default:
- **SSHD (SSH Daemon):** Enabled for secure remote access
- **TELNET Server:** Enabled for 5250 console connectivity

**Note:** If you encounter any issues with these services, you can manually start them using the commands mentioned in the sections below.

### Default User Profile (PAC)
The default user **PAC** comes with the following special authorities enabled:

| Authority | Description |
|-----------|-------------|
| `*ALLOBJ` | All object authority |
| `*SECADM` | Security administrator |
| `*JOBCTL` | Job control |
| `*AUDIT` | Audit authority |
| `*IOSYSCFG` | I/O system configuration |
| `*SAVSYS` | Save system |
| `*SERVICE` | Service authority |
| `*SPLCTL` | Spool control |

### Security Configuration
- **SSH Password Authentication:** Disabled by default for enhanced security
- **Authentication Method:** SSH key-based authentication is required

### System Configuration
The following system values have been modified to support virtual device connections (such as 5250 console from IBM ACS tool):

- **QAUTOVRT:** Configured to allow automatic virtual device creation
- **QLMTDEVSSN:** Device session limits adjusted for virtual connections
- **QLMTSECOFR:** Security officer limits configured appropriately

These pre-configured settings ensure that your IBM i system is ready for immediate use with secure remote access and 5250 console connectivity.

---

## Phase 1: Deployment via Power Access Cloud

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

## Phase 2: The IPL and Boot Process

**Important Note:**

Once the status shows **Active**, an External IP address will be assigned to your machine.

**For IBM i-based VMs**, the system must undergo an Initial Program Load (IPL) and several internal boot-up sequences. It typically takes **30–40 minutes** for the status to turn "Active" before the SSHD (SSH Daemon) starts and the machine becomes accessible via the network.

---

## Phase 3: Initial Remote Access

**Steps:**

Once the status is active, you can establish your first connection using the default administrative profile.

1. **Retrieve IP:** Copy the External IP from the service details page.
2. **Connect via SSH:** Open your local terminal (PowerShell, Command Prompt, or Terminal) and run:
   ```bash
   ssh PAC@<your-external-ip>
   ```

**Troubleshooting:**

If you encounter a "Host key verification failed" error, run the following command to remove the old host key from the known hosts file:

```bash
ssh-keygen -R <your-external-ip>
```

**Note:** If you are connecting as the system root user, you may need to use `sudo`:

```bash
sudo ssh-keygen -R <your-external-ip>
```

If prompted for a password when using `sudo`, enter your system password.

---

## Accessing 5250 Console via SSH-Key Tunneling

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
sudo ssh -i <path-to-private-key> -L 50000:localhost:23 -L 2001:localhost:2001  -L 449:localhost:449 -L 8470:localhost:8470 -L 8471:localhost:8471 -L 8472:localhost:8472 -L 2007:localhost:2007 -L 8473:localhost:8473 -L 8474:localhost:8474 -L 8475:localhost:8475 -L 8476:localhost:8476 -L 2003:localhost:2003 -L 2002:localhost:2002 -L 2006:localhost:2006 -L 2300:localhost:2300 -L 2323:localhost:2323 -L 3001:localhost:3001 -L 3002:localhost:3002 -L 2005:localhost:2005  -o ExitOnForwardFailure=yes -o ServerAliveInterval=15 -o ServerAliveCountMax=3 PAC@<external-ip>
```

**Note:**
- Some ports have restricted access, so it is recommended to use `sudo`.
- If prompted for a password when using `sudo`, enter your system password (not the PAC password).
- For future reference, visit https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-connect-ibmi#ssh-tunneling

**Parameters:**
- `i`: is used to specify path to the SSH private key file.
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

---

## Accessing Service Tools Menu from PAC User

This section describes how to access the IBM i System Service Tools (SST) menu using the PAC user account.

### Overview

The System Service Tools (SST) menu provides access to advanced system configuration and diagnostic tools. By default, the QSECOFR user profile is required to access SST. This guide walks you through enabling QSECOFR and configuring it to access the Service Tools menu.

### Prerequisites

- Active SSH connection to your IBM i system as the PAC user
- Access to the 5250 console via IBM ACS tool
- Console keypad functionality (F-keys) available in your 5250 session

**Note:** To show the keypad in the 5250 console, go to **View > Keypad** in the IBM ACS tool 5250 session window.

### Step 1: Modify QSECOFR User Profile

1. **Login as PAC user** via 5250 console
2. Run the following command to work with the QSECOFR user profile:
   ```
   WRKUSRPRF QSECOFR
   ```
3. **Press F21** on the console keypad to change the assistance level of the QSECOFR user from **Basic** to **Intermediate**
4. **Select option 2** to modify the QSECOFR user profile
5. **Enable the user** and **set a password** for QSECOFR

### Step 2: Configure DST Password

1. **Logout** of the PAC account
2. **Login as QSECOFR** using the password you just set
3. Run the following command to set the DST (Dedicated Service Tools) password to default:
   ```
   CHGDSTPWD PASSWORD(*DEFAULT)
   ```

### Step 3: Access Service Tools Menu

1. Run the following command to start System Service Tools:
   ```
   STRSST
   ```
2. At the SST sign-on screen, enter:
   - **User ID:** `QSECOFR`
   - **Password:** `QSECOFR` (default password) <!-- pragma: allowlist secret -->
3. You will receive a prompt indicating that the password has expired 
4. **Press F9** on the console keypad to change or set a new password
5. Follow the prompts to set a new secure password
6. Once the password is changed, you will be able to access the Service Tools menu

### Step 4: Future Access

After completing the initial setup, you can access the Service Tools menu directly from the PAC user account:

1. **Login as PAC user** (no need to switch to QSECOFR)
2. Run the command:
   ```
   STRSST
   ```
3. At the SST sign-on screen, enter:
   - **User ID:** `QSECOFR`
   - **Password:** The new password you set in Step 3 <!-- pragma: allowlist secret -->

You will now have access to the Service Tools menu without needing to logout and login as QSECOFR.

### Important Notes

- **Security:** Keep the QSECOFR password secure and change it regularly
- **Password Expiration:** If the password expires, you will need to repeat Step 3 to reset it
- **Assistance Level:** The assistance level change (Basic to Intermediate) provides access to additional system configuration options
- **Console Keypad:** To access the keypad in your 5250 console session, navigate to **View > Keypad** in the 5250 session window menu bar.

---