import "../App.css";
import React from "react";
import Login from "./Login";
import "bootstrap/dist/css/bootstrap.min.css";
import RequestList from "./RequestList";
import ServicesAdmin from "./Services-admin";
import NewRequest from "./PopUp/NewRequest";
import Dashboard from "./Dashboard";
import Register from "./Register";
import TnCRoute from "./PrivateRoute/TnCRoute";
import AuthRoute from "./PrivateRoute/AuthRoute";
import HeaderNav from "./Header";
import { Routes, Route } from "react-router-dom";
import UserService from "../services/UserService";
import "../App.css";
import Catalogs from "./Catalogs";
import CatalogsAdmin from "./Catalogs-admin";
import Users from "./Users";
import Events from "./Events";
import Keys from "./Keys";
import { Theme } from "@carbon/react";
import Feedbacks from "./Feedbacks";

const App = () => {
  const auth = UserService.isLoggedIn();
  const isAdmin = UserService.isAdminUser();

  const RouterClass = () => {
    return (
      <Routes>
        {!isAdmin && (
          <Route path="/" element={<TnCRoute Component={Dashboard} />} />
        )}

        {isAdmin && (
          <Route path="/" element={<TnCRoute Component={RequestList} />} />
        )}

        <Route path="/login" element={<AuthRoute Component={Login} />} />
        <Route path="/register" Component={Register} />

        {!isAdmin && (
          <Route
            path="/dashboard"
            element={<TnCRoute Component={Dashboard} />}
          />
        )}
        {!isAdmin && (
          <Route path="/catalogs" element={<TnCRoute Component={Catalogs} />} />
        )}

        {isAdmin && (
          <Route
            path="/requests"
            element={<TnCRoute Component={RequestList} />}
          />
        )}
        {isAdmin && (
          <Route
            path="/request/:id"
            element={<TnCRoute Component={NewRequest} />}
          />
        )}
        {isAdmin && (
          <Route
            path="/catalogs-admin"
            element={<TnCRoute Component={CatalogsAdmin} />}
          />
        )}
        {isAdmin && (
          <Route path="/users" element={<TnCRoute Component={Users} />} />
        )}
        {isAdmin && (
          <Route path="/keys" element={<TnCRoute Component={Keys} />} />
        )}
        {isAdmin && (
          <Route
            path="/services-admin"
            element={<TnCRoute Component={ServicesAdmin} />}
          />
        )}
        {isAdmin && (
          <Route path="/events" element={<TnCRoute Component={Events} />} />
        )}
        {isAdmin && (
          <Route
            path="/feedbacks"
            element={<TnCRoute Component={Feedbacks} />}
          />
        )}
      </Routes>
    );
  };
  if (auth === true && window.location.pathname === "/login") {
    window.location.href = window.location.href.replace("/login", "");
    return;
  }
  if (
    (auth === false || !isAdmin) &&
    [
      "/requests",
      "/catalogs-admin",
      "/users",
      "/services-admin",
      "/events",
      "/keys",
      "/feedbacks",
    ].includes(window.location.pathname)
  ) {
    window.location.href = "/login";
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
