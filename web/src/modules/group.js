import { SUCCESS_SUFFIX } from "redux-axios-middleware";
import UserService from "../services/UserService";

const LIST_GROUP = "LIST_GROUP";
const NEW_REQUEST = "NEW_REQUEST";

const groupReducer = (state = [], action) => {
  switch (action.type) {
    case LIST_GROUP + SUCCESS_SUFFIX:
      return action.payload.data;

    default:
      return state;
  }
};

export default groupReducer;

export const getGroup = (id) => ({
  type: LIST_GROUP,
  payload: {
    request: {
      url: `/pac-go-server/group/${id}`,
    },
  },
});

export const newRequest = (group) => {
    const username = UserService.getUsername();
    console.log(`${username} added the request ${group.name}`);
  
    const requestData = {
      justification: group.justification,
    };
  
    return {
      type: NEW_REQUEST,
      payload: {
        request: {
          url: `/pac-go-server/group/${group.id}/request`,
          method: "POST",
          data: requestData,
        },
      },
    };
  };