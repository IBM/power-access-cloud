import React, { Component } from "react";
import {
  Navbar,
  Button,
  Nav,
} from "react-bootstrap";
import Container from "react-bootstrap/Container";

import UserService from "../services/UserService";

// import Home from './Home';
// import Contact from './Contact';
// import About from './About';

export default class NavbarComp extends Component {
  render() {
    return (
      <Navbar bg="primary" variant="dark" sticky="top">
        <Container>
          <Navbar.Brand href="/">Power Access Cloud</Navbar.Brand>
          <Navbar.Toggle aria-controls="basic-navbar-nav" />
          <Navbar.Collapse id="basic-navbar-nav">
            <Nav className="me-auto">
              <Nav.Link href="/groups">Groups</Nav.Link>
              <Nav.Link href="/requests">Requests</Nav.Link>
            </Nav>
          </Navbar.Collapse>
          <Navbar.Collapse className="justify-content-end">
            <Navbar.Text className="mr-2">
              Signed in as: {UserService.getName()}({UserService.getUsername()})
            </Navbar.Text>
            <Button variant="danger" style={{ marginLeft: "10px" }}>Logout</Button>
          </Navbar.Collapse>
          <Nav.Link href="/about" className="text-info" style={{ marginLeft: "10px" }} >About</Nav.Link>
       </Container>
      </Navbar>
    );
  }
}
