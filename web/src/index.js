import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.scss';
import App from './components/App';
import reportWebVitals from './reportWebVitals';
import UserService from "./services/UserService";
import { Provider } from 'react-redux';
import { store } from './store/store';
import { GlobalTheme } from '@carbon/react';

const renderApp = () => {
  const root = ReactDOM.createRoot(document.getElementById('root'));
  root.render(
    <React.StrictMode>
      <Provider store={store}>
        <GlobalTheme theme={"g100"}>
          <App />
        </GlobalTheme>
      </Provider>
    </React.StrictMode>
  );
}

UserService.initKeycloak(renderApp);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
