import React, { useState } from "react";
import { UserAvatar } from "@carbon/icons-react";
import {
  Header,
  HeaderContainer,
  HeaderName,
  HeaderMenuButton,
  HeaderGlobalAction,
  HeaderGlobalBar,
  SideNav,
  SideNavItems,
  SideNavLink,
} from "@carbon/react";
import ProfileSection from "./Profile";
import UserService from "../services/UserService";

const HeaderNav = () => {
  const isAdmin = UserService.isAdminUser();
  const [showProfile, setShowProfile] = useState(false);
  return (
    <HeaderContainer
      render={({ isSideNavExpanded, onClickSideNavExpand }) => (
        <>
          <p>welcome</p>
          <Header aria-label="IBM Platform Name">
            <HeaderMenuButton
              aria-label={isSideNavExpanded ? "Close menu" : "Open menu"}
              isCollapsible
              onClick={onClickSideNavExpand}
              isActive={isSideNavExpanded}
              aria-expanded={isSideNavExpanded}
            />
            <HeaderName href="/" prefix="Power">
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
              {showProfile && <ProfileSection />}
            </HeaderGlobalBar>
            <SideNav
              aria-label="Side navigation"
              expanded={isSideNavExpanded}
              onOverlayClick={onClickSideNavExpand}
              onSideNavBlur={onClickSideNavExpand}
              isFixedNav={true}
              isChildOfHeader={false}
              style={{
                marginTop: "47px",
              }}
            >
              <SideNavItems>
                <SideNavLink href="/">Home</SideNavLink>
                <SideNavLink href="/groups">Groups</SideNavLink>
                <SideNavLink href="/requests">Request</SideNavLink>
                <SideNavLink href="/keys">Keys</SideNavLink>
                <SideNavLink href="/catalogs">Catalogs</SideNavLink>
                <SideNavLink href="/services">Services</SideNavLink>
                {isAdmin && <SideNavLink href="/users">Users</SideNavLink>}
                {isAdmin && <SideNavLink href="/events">Events</SideNavLink>}
              </SideNavItems>
            </SideNav>
          </Header>
        </>
      )}
    />
  );
};

export default HeaderNav;
