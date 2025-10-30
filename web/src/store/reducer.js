import { USER_LOGIN } from "./actionConstants";

const initialState = {
    context:{
      // Is the user logged in?
      isLoggedIn: false,
      // Logged in user
      user: null,
    }
  }
  
const rootReducer = (state = initialState, action) => {
  switch (action.type) {
    case USER_LOGIN:
      return {...state, context: action.context};
    default:
      return state;
  }
};
  
export default rootReducer;
