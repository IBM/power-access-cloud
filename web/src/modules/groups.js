import { SUCCESS_SUFFIX } from "redux-axios-middleware";
import UserService from "../services/UserService";

const LIST_GROUPS = "LIST_GROUPS";
const DELETE_GROUP = "DELETE_GROUP";

const groupsReducer = (state = [], action) => {
  switch (action.type) {
    case LIST_GROUPS + SUCCESS_SUFFIX:
      return action.payload.data;

    default:
      return state;
  }
};

export default groupsReducer;

export const allGroups = () => ({
  type: LIST_GROUPS,
  payload: {
    request: {
      url: "/pac-go-server/groups",
    },
  },
});

export const deleteGroup = (group) => {
  console.log(`${UserService.getUsername()} exits the group ${group.name}`);
  return {
    type: DELETE_GROUP,
    payload: {
      group,
      request: {
        url: `/pac-go-server/groups/${group.id}`,
        method: "DELETE",
      },
    },
  };
};


