import { Route, Routes } from "react-router-dom";
// import BookDetails from "./BookDetails";
// import BookForm from "./BookForm";
// import GroupList from "./GroupList";
// import RequestList from "./RequestList";
// import NoMatch from "./NoMatch";
// import RolesRoute from "./RolesRoute";
// import SecretBooks from "./SecretBooks";
import NavbarComp from "./Navbar";
import GroupList from "./GroupList";
import RequestList from "./RequestList";
import NewRequest from "./NewRequest";
import About from "./About";

const Pac = () => (
  <>
    <NavbarComp />
    <Routes>
    <Route exact path="/" element={<About />} />
      <Route exact path="/groups" element={<GroupList />} />
      <Route exact path="/requests" element={<RequestList />} />
      <Route exact path="/request/:id" element={<NewRequest />} />
      <Route exact path="/about" element={<About />} />
    </Routes>
  </>
);

export default Pac;
