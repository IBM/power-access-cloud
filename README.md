# IBM&reg; Power&reg; Access Cloud

IBM&reg; Power&reg; Access Cloud is a self-service portal that offers Open Source developers, ISVs, and customers a catalog of Power Hardware resources for a limited time. It enables users to develop, test, and optimize software on Power Architecture. With IBM&reg; Power&reg; Access Cloud, users can browse the catalog and select the desired Power Hardware resources based on their requirements. This self-service model eliminates the need for approvals, reduces manual intervention, and removes infrastructure barriers.

IBM&reg; Power&reg; Access Cloud aims to foster collaboration, innovation, and adoption of Power Architecture, simplifying the exploration and development of Power-based applications from a comprehensive catalog of available resources.

## Implementation Details

In the implementation of IBM&reg; Power&reg; Access Cloud, several key components have been utilized to provide a comprehensive and user-friendly experience for Open Source developers, ISVs, and customers. These include:

1. IBM Cloud Power Virtual Server: The IBM&reg; Power&reg; Access Cloud solution leverages the IBM Cloud Power Virtual Server as a foundational infrastructure to provide access to Power Systems architecture. This cloud-based infrastructure-as-a-service offering ensures reliable and scalable Power Hardware resources for users.

2. Self-Service Portal: IBM&reg; Power&reg; Access Cloud features a self-service portal that allows users to log in using their GitHub and IBM ID credentials. This portal serves as the interface through which users can access and manage their Power Hardware resources.

3. Access Request Management: Within the self-service portal, users can request access to the existing groups or quotas that control resource allocation. This helps in managing and controlling the availability of Power Hardware resources based on predefined limits and user requirements.

4. Catalog Browsing and Deployment: The self-service portal provides users with the ability to browse and deploy services from the existing catalogs. The current implementation focuses on a catalog of virtual machines, offering users a range of pre-configured VM options to choose from.

5. Auto Expiry and Approval Process: Deployed services within the IBM&reg; Power&reg; Access Cloud environment are set with auto expiry, ensuring efficient resource utilization. Users can request extensions for the service expiry period based on an approval process. This ensures proper resource management and fair distribution among users.

## Architecture

![image](https://github.ibm.com/pac/power-access-cloud/assets/512917/3d49f177-503f-42cb-8bb0-2455cded7370)

Please refer to: [devguide](api/internal/README.md) for creating setup locally.