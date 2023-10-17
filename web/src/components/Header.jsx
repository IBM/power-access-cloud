import React, { useState } from "react";
import { Link } from 'react-router-dom';
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

const MenuLink = (props) => {
  const { url, label } = props;
  return (
      <Link className="SideNavLink" to={url}>
        <SideNavLink>{label}</SideNavLink>
      </Link>
  );
}

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
                <MenuLink url="/" label="Home" />
                <MenuLink url="/catalogs" label="Catalogs" />
                <MenuLink url="/services" label="Services" />
                <MenuLink url="/groups" label="Groups" />
                <MenuLink url="/requests" label={isAdmin ? "Requests" : "My Requests"} />
                <MenuLink url="/keys" label="Keys" />
                {isAdmin && <MenuLink url="/users" label="Users" />}
                {isAdmin && <MenuLink url="/events" label="Events" />}
              </SideNavItems>
            </SideNav>
          </Header>
        </>
      )}
    />
  );
};

export default HeaderNav;
