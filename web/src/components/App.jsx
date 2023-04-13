// import logo from './logo.svg';
import '../App.css';
import { Provider } from "react-redux";
import { BrowserRouter } from "react-router-dom";
import RenderOnAnonymous from "./RenderOnAnonymous";
import RenderOnAuthenticated from "./RenderOnAuthenticated";
import Welcome from "./Welcome";
import 'bootstrap/dist/css/bootstrap.min.css';
import Pac from './Pac';

const App = ({ store }) => (
  <Provider store={store}>
    <BrowserRouter>
        <div className="container">
        <RenderOnAnonymous>
        <Welcome/>
          {console.log("Welcome to RenderOnAnonymous")}
        </RenderOnAnonymous>
        <RenderOnAuthenticated>
          {console.log("Welcome to RenderOnAuthenticated")}
          <Pac/>
        </RenderOnAuthenticated>
        </div>
    </BrowserRouter>
  </Provider>
);

export default App;
