import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './components/App';
import reportWebVitals from './reportWebVitals';
import UserService from "./services/UserService";
import { applyMiddleware, createStore } from "redux";
import axiosMiddleware from "redux-axios-middleware";
import axios from "axios";
import thunk from "redux-thunk";
import rootReducer from "./modules";

// HTTP

const _axios = axios.create();
_axios.interceptors.request.use((config) => {
  if (UserService.isLoggedIn()) {
    const cb = () => {
      config.headers.Authorization = `Bearer ${UserService.getToken()}`;
      return Promise.resolve(config);
    };
    return UserService.updateToken(cb);
  }
});

// REDUX STORE

const _middleware = applyMiddleware(thunk, axiosMiddleware(_axios));
const store = createStore(rootReducer, _middleware);


const renderApp = () => {
  const root = ReactDOM.createRoot(document.getElementById('root'));
  root.render(
    <React.StrictMode>
      <App store={store}/>
    </React.StrictMode>
  );
}

UserService.initKeycloak(renderApp);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
