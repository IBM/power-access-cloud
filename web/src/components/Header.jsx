import React, { useState } from "react";
import { Link } from 'react-router-dom';
import { UserAvatar } from "@carbon/icons-react";
import Feedback from "./PopUp/Feedback";

import {
  Header,
  HeaderContainer,
  HeaderName,
  HeaderMenuButton,
  HeaderGlobalAction,
  HeaderNavigation,
  HeaderMenuItem,
  HeaderGlobalBar,
  SideNav,
  SideNavItems,
  SideNavLink,
} from "@carbon/react";
import ProfileSection from "./Profile";
import UserService from "../services/UserService";
const BUTTON_FEEDBACK = "BUTTON_FEEDBACK";
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
  const [actionProps, setActionProps] = useState("");
  const action={
    key: BUTTON_FEEDBACK,
      label: "Feedback"
  }
  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_FEEDBACK && (
          <Feedback
            
            setActionProps={setActionProps}
            
          />
        )}
        </React.Fragment>)}
  return (
    <HeaderContainer
    
      render={({ isSideNavExpanded, onClickSideNavExpand }) => (
        <>
           {renderActionModals()}
          <Header aria-label="">
          {isAdmin&& <HeaderMenuButton
              aria-label={isSideNavExpanded ? "Close menu" : "Open menu"}
              isCollapsible
              onClick={onClickSideNavExpand}
              isActive={isSideNavExpanded}
              aria-expanded={isSideNavExpanded}
            />}
            <HeaderName as={Link} to="/" prefix="">
             Power Access Cloud
            </HeaderName>
            {!isAdmin&&<HeaderNavigation aria-label="">
              <HeaderMenuItem as={Link} to="catalogs">Catalog</HeaderMenuItem>
              <HeaderMenuItem as={Link} to="https://github.com/PDeXchange/pac-support/blob/main/docs/FAQ.md">FAQ</HeaderMenuItem>
              <HeaderMenuItem onClick={() => setActionProps(action)}>Feedback</HeaderMenuItem>
            </HeaderNavigation>}
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
            {isAdmin&&<SideNav
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
                <MenuLink url="/" label={isAdmin ? "Requests" : "Dashboard"} />
                
                <MenuLink url={isAdmin ? "/catalogs-admin" : "/catalogs"} label="Catalog" />
                <MenuLink url="/feedback" label="Feedback" />
                {isAdmin && <MenuLink url="/services-admin" label="Services" />}
                {isAdmin && <MenuLink url="/keys" label="Keys" />}
                {isAdmin && <MenuLink url="/users" label="Users" />}
                {isAdmin && <MenuLink url="/events" label="Events" />}
              </SideNavItems>
            </SideNav>
            }
          </Header>
        </>
      )}
    />
  );
};

export default HeaderNav;
