import React, { useState, useEffect } from "react";
import { Modal, Theme, Dropdown } from "@carbon/react";
import { allGroups, updateUserGroup } from "../../services/request";

const ModifyUserGroups = ({ user, setActionProps, response, fetchData }) => {
  const [availableGroups, setAvailableGroups] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const fetchGroups = async () => {
      try {
        const { payload } = await allGroups();
        const groupItems = payload.map((group) => ({
          id: group.id,
          label: group.name,
        }));
        setAvailableGroups(groupItems);

        // Pre-select user's current group (first one if multiple exist)
        const userGroups = Array.isArray(user.groups) ? user.groups : [];
        if (userGroups.length > 0) {
          const currentGroup = groupItems.find((group) =>
            group.label === userGroups[0]
          );
          setSelectedGroup(currentGroup);
        }
        setLoading(false);
      } catch (error) {
        console.log("Failed to fetch groups:", error);
        setLoading(false);
      }
    };

    fetchGroups();
  }, [user.groups]);

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;

    if (!selectedGroup) {
      title = "User group update failed.";
      message = "Please select a group.";
      errored = true;
      response(title, message, errored);
      setActionProps("");
      return;
    }

    setSubmitting(true);

    try {
      const { type, payload } = await updateUserGroup(user.id, selectedGroup.id);
      
      if (type === "API_ERROR") {
        title = "User group update failed.";
        message = payload.response?.data?.error || "An error occurred";
        errored = true;
      } else {
        title = "User group updated successfully.";
        message = `User ${user.username} has been moved to group "${selectedGroup.label}".`;
        await fetchData(); // Refresh the user list
      }
    } catch (error) {
      console.log("Update failed: ", error);
      title = "User group update failed.";
      message = error.message || "An unexpected error occurred";
      errored = true;
    } finally {
      setSubmitting(false);
    }

    response(title, message, errored);
    setActionProps("");
  };

  return (
    <Theme theme="g10">
      <Modal
        modalLabel="Modify User Group"
        modalHeading={`Change group for user "${user.username}"`}
        size="lg"
        className="modify-user-groups-modal"
        onRequestClose={() => {
          setActionProps("");
        }}
        onRequestSubmit={() => {
          onSubmit();
        }}
        open={true}
        primaryButtonText={submitting ? "Saving..." : "Save"}
        secondaryButtonText={"Cancel"}
        primaryButtonDisabled={loading || !selectedGroup || submitting}
        secondaryButtonDisabled={submitting}
      >
        <div className="modify-user-groups-modal__content">
          <div className="mb-3">
            <p><strong>User ID:</strong> {user.id}</p>
            <p><strong>Name:</strong> {user.firstname} {user.lastname}</p>
            <p><strong>Email:</strong> {user.email}</p>
            <br />
            <Dropdown
              id="user-group-select"
              titleText="Select Group"
              label="Select a group for this user"
              items={availableGroups}
              itemToString={(item) => (item ? item.label : "")}
              selectedItem={selectedGroup}
              onChange={({ selectedItem }) => {
                setSelectedGroup(selectedItem);
              }}
              disabled={loading}
              size="lg"
            />
            <p style={{ marginTop: '1rem', fontSize: '0.875rem', color: '#525252' }}>
              Note: This will remove the user from all current groups and add them to the selected group.
            </p>
          </div>
        </div>
      </Modal>
    </Theme>
  );
};

export default ModifyUserGroups;
