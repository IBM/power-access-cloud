import React, { useState, useEffect, useRef } from "react";
import { Link } from "react-router-dom";
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
import ToastNotify from "./utils/ToastNotify";
import "../styles/carbon-override.scss"

const BUTTON_FEEDBACK = "BUTTON_FEEDBACK";
const MenuLink = (props) => {
  const { url, label } = props;
  return (
    <Link className="SideNavLink" to={url}>
      <SideNavLink>{label}</SideNavLink>
    </Link>
  );
};

const HeaderNav = ({ onSideNavToggle }) => {
  const isAdmin = UserService.isAdminUser();
  const [showProfile, setShowProfile] = useState(false);
  const [actionProps, setActionProps] = useState("");
  const [isSideNavExpanded, setIsSideNavExpanded] = useState(false);
  const sideNavRef = useRef(null);

  // notify
  const [notifyKind, setNotifyKind] = useState("");
  const [title, setTitle] = useState("");
  const [message, setMessage] = useState("");

  // Handle click outside to close side nav
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (isSideNavExpanded && sideNavRef.current && !sideNavRef.current.contains(event.target)) {
        // Check if click is not on the hamburger button
        const hamburgerButton = document.querySelector('.cds--header__menu-toggle');
        if (hamburgerButton && !hamburgerButton.contains(event.target)) {
          setIsSideNavExpanded(false);
          if (onSideNavToggle) {
            onSideNavToggle(false);
          }
        }
      }
    };

    if (isSideNavExpanded) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isSideNavExpanded, onSideNavToggle]);

  const action = {
    key: BUTTON_FEEDBACK,
    label: "Feedback",
  };

  const handleFeedBackResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
  };

  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_FEEDBACK && (
          <Feedback
            setActionProps={setActionProps}
            response={handleFeedBackResponse}
          />
        )}
      </React.Fragment>
    );
  };
  const toggleSideNav = () => {
    const newState = !isSideNavExpanded;
    setIsSideNavExpanded(newState);
    if (onSideNavToggle) {
      onSideNavToggle(newState);
    }
  };

  return (
    <HeaderContainer
      render={() => (
        <>
          {renderActionModals()}
          <Header aria-label="">
            {isAdmin && (
              <HeaderMenuButton
                aria-label={isSideNavExpanded ? "Close menu" : "Open menu"}
                isCollapsible
                onClick={toggleSideNav}
                isActive={isSideNavExpanded}
                aria-expanded={isSideNavExpanded}
              />
            )}
            <HeaderName as={Link} to="/" prefix="">
              IBM&reg; Power&reg; Access Cloud
            </HeaderName>
            {!isAdmin && (
              <HeaderNavigation aria-label="">
                <HeaderMenuItem as={Link} to="catalogs">
                  Catalog
                </HeaderMenuItem>
                <HeaderMenuItem
                  as={Link}
                  to="https://github.com/IBM/power-access-cloud/blob/main/support/docs/FAQ.md"
                >
                  FAQ
                </HeaderMenuItem>
                <HeaderMenuItem onClick={() => setActionProps(action)}>
                  Feedback
                </HeaderMenuItem>
                <ToastNotify
                  title={title}
                  subtitle={message}
                  kind={notifyKind}
                  setTitle={setTitle}
                />
              </HeaderNavigation>
            )}
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
            {isAdmin && (
              <div ref={sideNavRef}>
                <SideNav
                  aria-label="Side navigation"
                  expanded={isSideNavExpanded}
                  onOverlayClick={toggleSideNav}
                  isFixedNav={true}
                  isChildOfHeader={false}
                  style={{
                    marginTop: "47px",
                  }}
                >
                <SideNavItems>
                  <MenuLink
                    url="/"
                    label={isAdmin ? "Requests" : "Dashboard"}
                  />

                  <MenuLink
                    url={isAdmin ? "/catalogs-admin" : "/catalogs"}
                    label="Catalog"
                  />
                  <MenuLink url="/feedbacks" label="Feedback" />
                  {isAdmin && (
                    <MenuLink url="/services-admin" label="Services" />
                  )}
                  {isAdmin && <MenuLink url="/keys" label="Keys" />}
                  {isAdmin && <MenuLink url="/users" label="Users" />}
                  {isAdmin && <MenuLink url="/events" label="Events" />}
                </SideNavItems>
              </SideNav>
              </div>
            )}
          </Header>
        </>
      )}
    />
  );
};

export default HeaderNav;
