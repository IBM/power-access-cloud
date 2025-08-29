# Frequently Asked Questions (FAQ)

Welcome to the FAQ page for IBM&reg; Power&reg; Access Cloud. Below are some common questions and their answers.

---

## 1. What is IBM&reg; Power&reg; Access Cloud?

**Answer:**  
IBM&reg; Power&reg; Access Cloud is a self-service portal that offers Open Source developers, ISVs, and customers a catalog of Power Hardware resources for a limited time. It enables users to develop, test, and optimize software on Power Architecture. With IBM&reg; Power&reg; Access Cloud, users can browse the catalog and select the desired Power Hardware resources based on their requirements. This self-service model eliminates the need for approvals, reduces manual intervention, and removes infrastructure barriers.

---

## 2. How do I Register with IBM&reg; Power&reg; Access Cloud?

**Answer:**  
Below steps provides the quick guide to get started with IBM&reg; Power&reg; Access Cloud
1. Go to our **[IBM&reg; Power&reg; Access Cloud Portal](https://console.developonpower.ibm.com)**.
2. Click the **Register** button.
3. Fill in the text area on how do you plan to use the IBM&reg; Power&reg; Access Cloud service.
4. Read and Accept our **IBM&reg; Power&reg; Access Cloud terms and conditions**.
5. Click the **Submit** button.

---

## 3. How can I Quickly Get Started with IBM&reg; Power&reg; Access Cloud?

**Answer:**  
Below steps provides the quick guide on how to get started with IBM&reg; Power&reg; Access Cloud

1. Go to our **[IBM&reg; Power&reg; Access Cloud Portal](https://console.developonpower.ibm.com)**.
2. Click the **Login** button.
3. Sign in to your account using either **Github** or **IBMid Identity Provider**.
4. On successful login, you will be landing in to the IBM&reg; Power&reg; Access Cloud Dashboard.
5. Make sure you're part of a group and Status is **Approved** (If still **Pending**, please wait for the admin to approve or reach out to us through email as mentioned in **Support**).
6. Add a new key under **My keys** section (Please Refer ***Q5*** for more info).
7. Deploy a service by clicking on **Go to Catalog** button under **My services** and selecting the desired instance type.
8. Wait for some time till the service is deployed successfully with **Status: Active**.
9. Once Active, you will be able to access the instance using SSH (Please Refer ***Q7*** for more info)

---

## 4. What is a group and How can I request for more resources?

**Answer:**
A Group is a logical unit which controls resource allocation by allocating the maximum vCPU and memory quotas. By default, all new users are added to the Bronze group which includes .5 vCPU and 8 GB of memory. 

To request for more CPU and memory resources, you can submit a "Request" by clicking on **Upgrade** and selecting a desired group along with the justification for upgrading.
Please wait for some time until the request is approved by the **Admin** (You will be notified via email on successful approval/ rejection).
Upon successful approval, you will now be part of the **requested group**

***Note***:  You can only be a member of one group at a time.

---

## 5. What are the need for Keys and How do I generate a Public Key?

**Answer:**
Keys enable **Secure** and **Passwordless Authentication** to your services so that only you/ desired people will be able to access it. You can add a key by clicking on **Add Key**. Also you can add/create multiple keys to enable secure access from multiple systems.

To generate a Public Key, follow the below steps
#### On **Linux** or **macOS**:
1. **Open a terminal** window.
2. Run the following command to generate a new SSH key pair:
   ```bash
   ssh-keygen -t rsa -b 4096 -C "your_email@example.com"```
3. When prompted for a file location to save the key, press Enter to accept the default location:
    ```bash
    /home/youruser/.ssh/id_rsa  (Linux) or /Users/youruser/.ssh/id_rsa (macOS)```

4. When prompted, enter a passphrase for added security, or press Enter to leave it empty.
5. Your new SSH key pair will be generated, and the public key (id_rsa.pub) will be ready for use.

---

## 6. What are Catalogs?

**Answer:**
Catalog provides rich set of flavours of multiple predefined instance types from which you can deploy a service.

---

## 7. How to provision a service and How do I access my service after it's deployed?

**Answer:**
You can provision a service by going to the Catalog and selecting the required instance type.
***Note***: If you do not have sufficient quota, a warning message will be popped up saying **Insufficient Quota**. You can either **Go to Dashboard** to free up the unused services or upgrade your existing group to get more CPU and memory quota.

Once the service is deployed and Status is set to **Active**, you can access it by doing ssh using the below command.
```bash
ssh -i ~/.ssh/id_rsa root@<ip> # Replace <ip> with actual IP available in 'Access Information'
```
***Note***: Post creation of service and its available, please wait for few minutes before accessing.

---

## 8. How do I contact support?

**Answer:**  
If you're having issues, you can reach out to us at **PowerACL@ibm.com**, or open an issue on our [GitHub Issues Page](https://github.com/PDEXchange/pac-support). Our team will get back to you as soon as possible.

---

## 9. Where can I find the source code?

**Answer:**  
The source code for IBM&reg; Power&reg; Access Cloud is hosted on GitHub. You can access it here: [IBM&reg; Power&reg; Access Cloud](https://github.com/PDeXchange/pac) and [IBM&reg; Power&reg; Access Cloud-UI](https://github.com/PDeXchange/pac-ui).

---

## 10. How can I report a bug or suggest a feature?

**Answer:**  
To report bugs or request new features, please visit [GitHub Issues Page](https://github.com/PDeXchange/pac) on our GitHub repository.

---

## 11. How do I report a security vulnerability?

**Answer:**  
If you discover a security vulnerability in this project, we encourage you to report it responsibly. You can send an email to our security team at **PowerACL@ibm.com** with Subject having **[Security]** prefixed, and we will address it as quickly as possible.

We take security seriously and will work with you to resolve any issues as swiftly as possible.

---