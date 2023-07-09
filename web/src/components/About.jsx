import React from "react";
import {
  Button,
  Tabs,
  Tab,
  TabList,
  TabPanels,
  TabPanel,
  Grid,
  Column,
} from "@carbon/react";

const About = () => {
  return (
    <Grid className="landing-page" fullWidth>
      <Column lg={16} md={8} sm={4} className="landing-page__banner">
        <h1 className="landing-page__heading">
          Power Access Cloud
        </h1>
      </Column>
      <Column lg={16} md={8} sm={4} className="landing-page__r2">
        <Tabs defaultSelectedIndex={0}>
          <TabList className="tabs-group" aria-label="Tab navigation">
            <Tab>About</Tab>
            <Tab>Step 1</Tab>
            <Tab>Step 2</Tab>
            <Tab>Step 3</Tab>
            <Tab>Step 4</Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              <Grid className="tabs-group-content">
                <Column
                  md={8}
                  lg={16}
                  sm={4}
                  className="landing-page__tab-content"
                >
                  <h1 className="landing-page__subheading">What is Power Access Cloud(PAC)?</h1>
                  <p className="landing-page__p">
                    PAC is a self-service portal that offers Open Source developers access to Power Hardware resources for a limited time. It aims to simplify the exploration, development, and optimization of software on Power Architecture.
                  </p>
                  <h2 className="landing-page__subheading">Key Features:</h2>
                  <ol>
                    <li>
                      <p>
                        <strong>Self-Service Portal:</strong> Users can log in to the PAC self-service portal using their GitHub and IBM ID credentials. The portal serves as the interface to access and manage Power Hardware resources.
                      </p>
                    </li>
                    <li>
                      <p>
                        <strong>Catalog Browsing and Deployment:</strong> The self-service portal provides a catalog of available Power Hardware resources. Users can browse the catalog, select the desired resources, and deploy them according to their requirements.
                      </p>
                    </li>
                    <li>
                      <p>
                        <strong>Access Request Management:</strong> Users can request access to specific groups that control resource allocation. This helps manage and control the availability of Power Hardware resources based on predefined limits and user needs.
                      </p>
                    </li>
                    <li>
                      <p>
                        <strong>Auto Expiry and Approval Process:</strong> Deployed services within PAC are set with auto expiry to ensure efficient resource utilization. Users can request extensions for service expiration, which undergo an approval process to ensure fair distribution.
                      </p>
                    </li>
                  </ol>
                  <p>
                    PAC aims to foster collaboration, innovation, and adoption of Power Architecture by offering a comprehensive catalog of Power Hardware resources and enabling a self-service model. It eliminates the need for manual approvals, reduces infrastructure barriers, and provides a user-friendly experience for developers.
                  </p>
                </Column>
              </Grid>
            </TabPanel>
            <TabPanel>
              <Grid className="tabs-group-content">
                <Column
                  lg={16}
                  md={8}
                  sm={4}
                  className="landing-page__tab-content"
                >
                  <h2 className="landing-page__subheading">Join the group</h2>
                  <p className="landing-page__p">
                    PAC allows users to join existing groups that control resource allocation and access permissions.
                    Navigate to the "Groups" section in the self-service portal.
                    Browse through the available groups and their descriptions to find one that aligns with your interests or requirements.
                  </p>
                </Column>
                <Column
                  lg={16}
                  md={8}
                  sm={4}
                  className="landing-page__tab-content"
                >
                  <h1 className="landing-page__subheading">Request Group Access:</h1>
                  <p className="landing-page__p">

                    Select the desired group and click on the "Request Access" button. Provide any additional information or justification for joining the group in the access request form.
                    Submit your request.
                  </p>

                </Column>
                <Column
                  lg={16}
                  md={8}
                  sm={4}
                  className="landing-page__tab-content"
                >
                  <h1 className="landing-page__subheading">Track Request Status:</h1>
                  <p className="landing-page__p">

                    Monitor the status of your group access request within the self-service portal.
                    You will be notified once your request is approved or rejected.
                    Submit your request.
                  </p>
                  <Button href="/groups">Join</Button> <Button href="/requests">Track</Button>
                </Column>
              </Grid>
            </TabPanel>
            <TabPanel>
              <Grid className="tabs-group-content">
                <Column
                  lg={16}
                  md={8}
                  sm={4}
                  className="landing-page__tab-content"
                >
                  <h2 className="landing-page__subheading">SSH keys</h2>
                  <p className="landing-page__p">
                    To access the Power Access Cloud service user needs to have ssh key to access the deployed service. If you don't have ssh keys, please generate one and upload it to the portal.
                  </p>
                  <Button href="/keys">Manage Keys</Button>
                </Column>
              </Grid>
            </TabPanel>
            <TabPanel>
              <Grid className="tabs-group-content">
                <Column
                  lg={16}
                  md={8}
                  sm={4}
                  className="landing-page__tab-content"
                >
                  <h2 className="landing-page__subheading">Catalogs</h2>
                  <p>
                    PAC offers a catalog of available Power Hardware resources for users to browse and select from. The catalog provides a range of options that users can choose based on their requirements. It simplifies the process of accessing and utilizing Power Architecture resources.
                  </p>
                  <h3 className="landing-page__subheading">Browsing the Catalog</h3>
                  <p>
                    Within the PAC self-service portal, users can navigate to the catalog section to explore the available Power Hardware resources. The catalog provides detailed information about each resource, including specifications, configurations, and any pre-configured settings.
                  </p>
                  <h3 className="landing-page__subheading">Selecting and Deploying Resources</h3>
                  <p>
                    Once users have identified the desired Power Hardware resources from the catalog, they can proceed to select and deploy them. The self-service portal provides a straightforward process for resource selection and deployment.
                  </p>
                  <p>
                    After selecting the catalog, users can initiate the deployment process. This will provision the selected Power Hardware resource and make it available for use within the designated environment.
                  </p>
                  <p>
                    It's important to note that the availability of specific resources within the catalog may vary based on factors such as quotas, group permissions, and overall resource availability within the PAC infrastructure.
                  </p>
                  <Button href="/catalogs">Catalogs</Button>
                </Column>
              </Grid>
            </TabPanel>
            <TabPanel>
              <Grid className="tabs-group-content">
                <Column
                  lg={16}
                  md={8}
                  sm={4}
                  className="landing-page__tab-content"
                >
                  <h2 className="landing-page__subheading">Deployed Services</h2>
                  <p>
                    Once users have selected and deployed Power Hardware resources from the catalog, these resources become deployed services within the PAC environment. Deployed services offer users the allocated Power Hardware resources for their development, testing, and optimization needs.
                  </p>
                  <h3 className="landing-page__subheading">Accessing Deployed Services</h3>
                  <p>
                    Users can access their deployed services through the PAC self-service portal. The portal provides access details and relevant information about the deployed services, such as connection instructions and credentials.
                  </p>
                  <p>
                    Depending on the specific Power Hardware resources deployed, users may connect to virtual machines or other resources within their allocated environment. They can use remote access tools or protocols to establish a connection and interact with the deployed services.
                  </p>
                  <h3 className="landing-page__subheading">Developing and Testing on Power Hardware</h3>
                  <p>
                    With the deployed services, users have the opportunity to develop, test, and optimize their software on the Power Architecture infrastructure. They can leverage the allocated Power Hardware resources to run their applications, perform testing, and evaluate the compatibility and efficiency of their software on Power Architecture.
                  </p>
                  <p>
                    Users can install and configure their development tools, libraries, and dependencies within the deployed services to create an environment conducive to their specific development needs.
                  </p>
                  <h3 className="landing-page__subheading">Monitoring and Managing Deployed Services</h3>
                  <p>
                    Throughout the development and testing process, users can monitor and manage their deployed services through the PAC self-service portal. They can track resource status, and perform necessary management actions, such as ask for extension for the service expiry, or terminating the deployed services.
                  </p>
                  <p>
                    It's important to note that deployed services within PAC are set with an auto-expiry feature to ensure efficient resource utilization. Users should keep track of the expiration date and request extensions if required, following the approval process outlined in PAC.
                  </p>
                  <Button href="/services">Services</Button>
                </Column>
              </Grid>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </Column>
      <Column lg={16} md={8} sm={4} className="landing-page__r3">
        <Grid>
          <Column md={4} lg={4} sm={4}>
            <h3 className="landing-page__label">The Principles</h3>
          </Column>
          <Column md={4} lg={4} sm={4}>
            Open
          </Column>
          <Column md={4} lg={4} sm={4}>
            Collaboration
          </Column>
          <Column md={4} lg={4} sm={4}>
            Innovation
          </Column>
        </Grid>
      </Column>
    </Grid>
  );
};

export default About;
