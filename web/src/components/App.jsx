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
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import UserService from "../services/UserService";
import "../App.css";
import Keys from "./Keys";
import Catalogs from "./Catalogs";
import Services from "./Services";
import Users from "./Users";

const App = () => {
  const auth = UserService.isLoggedIn();
  const isAdmin = UserService.isAdminUser();

  // const navigate = useNavigate();
  const router = createBrowserRouter([
    {
      path: "/",
      element: <AuthRoute Component={About} />,
    },
    {
      path: "/login",
      element: <AuthRoute Component={Login} />,
    },
    {
      path: "/groups",
      element: <AuthRoute Component={GroupList} />,
    },
    {
      path: "/requests",
      element: <AuthRoute Component={RequestList} />,
    },
    {
      path: "/request/:id",
      element: <AuthRoute Component={NewRequest} />,
    },
    {
      path: "/about",
      element: <AuthRoute Component={About} />,
    },
    {
      path: "/keys",
      element: <AuthRoute Component={Keys} />,
    },
    {
      path: "/catalogs",
      element: <AuthRoute Component={Catalogs} />,
    },
    {
      path: "/services",
      element: <AuthRoute Component={Services} />,
    },
    isAdmin && {
      path: "/users",
      element: <AuthRoute Component={Users} />,
    },
  ]);
  if (auth === true && window.location.pathname === "/login") {
    window.location.href = window.location.href.replace("/login", "");
    return;
  }
  return (
    <React.Fragment>
      {auth === true && <HeaderNav />}
      <section className={auth ? "contentSection" : ""}>
        <RouterProvider router={router} />
      </section>
    </React.Fragment>
  );
};

export default App;
