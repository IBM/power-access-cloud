import "../App.css";
import React from "react";
import Login from "./Login";
import "bootstrap/dist/css/bootstrap.min.css";
import GroupList from "./GroupList";
import RequestList from "./RequestList";
import NewRequest from "./PopUp/NewRequest";
import About from "./About";
import Terms from "./Terms";
import AuthRoute from "./PrivateRoute/AuthRoute";
import TnCRoute from "./PrivateRoute/TnCRoute";
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
        <Route path="/" element={<TnCRoute Component={About} />} />
        <Route path="/login" element={<AuthRoute Component={Login} />} />
        <Route path="/terms" element={<AuthRoute Component={Terms} />} />
        <Route path="/groups" element={<TnCRoute Component={GroupList} />} />
        <Route path="/requests" element={<TnCRoute Component={RequestList} />} />
        <Route path="/request/:id" element={<TnCRoute Component={NewRequest} />} />
        <Route path="/about" element={<TnCRoute Component={About} />} />
        <Route path="/keys" element={<TnCRoute Component={Keys} />} />
        <Route path="/catalogs" element={<TnCRoute Component={Catalogs} />} />
        <Route path="/services" element={<TnCRoute Component={Services} />} />
        {isAdmin && <Route path="/users" element={<TnCRoute Component={Users} />} />}
        {isAdmin && <Route path="/events" element={<TnCRoute Component={Events} />} />}
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
