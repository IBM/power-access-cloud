import React, { useState } from "react";
import { UserAvatar } from "@carbon/icons-react";
import {
  Header,
  HeaderName,
  HeaderGlobalAction,
  HeaderGlobalBar,
  Theme,
  SideNav,
  SideNavItems,
  SideNavLink,
} from "@carbon/react";
import ProfileSection from "./Navbar";

const HeaderNav = () => {
  const [showProfile, setShowProfile] = useState(false);
  return (
    <>
      <>
        <div className="userDetails">
          <p>User1</p>
        </div>
        <div className="container">
          <Theme theme="g100">
            <Header aria-label="IBM Platform Name">
              <HeaderName href="#" prefix="Power">
                Access Cloud
              </HeaderName>
              <HeaderGlobalBar>
                <HeaderGlobalAction
                  aria-label="Profile"
                  onClick={() => {
                    setShowProfile(!showProfile);
                  }}
                >
                  <UserAvatar size="32" tabIndex="0" />
                </HeaderGlobalAction>
              </HeaderGlobalBar>
            </Header>
          </Theme>
        </div>
        {showProfile && <ProfileSection />}
      </>
      <SideNav
        isFixedNav={true}
        expanded={true}
        isChildOfHeader={false}
        aria-label="Side navigation"
        style={{
          marginTop: "47px",
          backgroundColor: "black",
          color: "white",
          fontSize: "1.5rem",
        }}
        className="sidebar-fixed"
      >
        <SideNavItems>
          <SideNavLink href="/groups">Groups</SideNavLink>
          <SideNavLink href="/requests">Request</SideNavLink>
          <SideNavLink href="/keys">Keys</SideNavLink>
          <SideNavLink href="/catalogs">Catalogs</SideNavLink>
          <SideNavLink href="/services">Services</SideNavLink>
          <SideNavLink href="/">About</SideNavLink>
        </SideNavItems>
      </SideNav>
    </>
  );
};

export default HeaderNav;
/*
For Future Use
<SideNavMenu title="L0 menu">
    <SideNavMenuItem href="https://www.carbondesignsystem.com/">
        L0 menu item
    </SideNavMenuItem>
</SideNavMenu>
*/
