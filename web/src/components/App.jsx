import "../App.css";
import React from "react";
import Login from "./Login";
import "bootstrap/dist/css/bootstrap.min.css";
import GroupList from "./GroupList";
import RequestList from "./RequestList";
import NewRequest from "./PopUp/NewRequest";
import About from "./About";
import AuthRoute from "./PrivateRoute/AuthRoute";
import HeaderNav from "./Header";
import {Routes , Route } from "react-router-dom";
import UserService from "../services/UserService";
import "../App.css";
import Keys from "./Keys";
import Catalogs from "./Catalogs";
import Services from "./Services";
import Users from "./Users";
import Events from "./Events";
import { Theme } from "@carbon/react";

const App = () => {
  const auth = UserService.isLoggedIn();
  const isAdmin = UserService.isAdminUser();

  const RouterClass = () => {
    return(
      <Routes>
      <Route path="/" element={<AuthRoute Component={About} />} />
      <Route path="/login" element={<AuthRoute Component={Login} />} />
      <Route path="/groups" element={<AuthRoute Component={GroupList} />} />
      <Route path="/requests" element={<AuthRoute Component={RequestList} />} />
      <Route path="/request/:id" element={<AuthRoute Component={NewRequest} />} />
      <Route path="/about" element={<AuthRoute Component={About} />} />
      <Route path="/keys" element={<AuthRoute Component={Keys} />} />
      <Route path="/catalogs" element={<AuthRoute Component={Catalogs} />} />
      <Route path="/services" element={<AuthRoute Component={Services} />} />
      {isAdmin && <Route path="/users" element={<AuthRoute Component={Users} />} />}
      {isAdmin && <Route path="/events" element={<AuthRoute Component={Events} />} />}
    </Routes>
    );
  }

  if (auth === true && window.location.pathname === "/login") {
    window.location.href = window.location.href.replace("/login", "");
    return;
  }
  return (
    <React.Fragment>
      <Theme theme="g90">{auth === true && <HeaderNav />} </Theme>
      <section className={auth ? "contentSection" : ""}>
        <RouterClass />
      </section>
    </React.Fragment>
  );
};

export default App;
