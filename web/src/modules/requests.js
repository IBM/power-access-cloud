import { SUCCESS_SUFFIX } from "redux-axios-middleware";
import UserService from "../services/UserService";

const LIST_REQUESTS = "LIST_REQUESTS";
const ADD_REQUEST = "ADD_REQUEST"

const groupsReducer = (state = [], action) => {
  switch (action.type) {
    case LIST_REQUESTS + SUCCESS_SUFFIX:
      return action.payload.data;

    default:
      return state;
  }
};

export default groupsReducer;

export const allRequests = () => ({
  type: LIST_REQUESTS,
  payload: {
    request: {
      url: "/pac-go-server/requests",
    },
  },
});

export const addRequest = (group) => {
  const username = UserService.getUsername();
  console.log(`${username} added the request ${group.name}`);

  const requestData = {
    group: group.name,
    requester: username,
  };

  return {
    type: ADD_REQUEST,
    payload: {
      request: {
        url: "/pac-go-server/request",
        method: "POST",
        data: requestData,
      },
    },
  };
};
