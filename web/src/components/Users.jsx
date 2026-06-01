import React, { useState, useEffect } from "react";
import { allUsers } from "../services/request";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
import {
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  TableToolbar,
  TableToolbarSearch,
  DataTableSkeleton,
  Tag,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";
import DeleteUser from "./PopUp/DeleteUser";
import ModifyUserGroups from "./PopUp/ModifyUserGroups";
import UserService from "../services/UserService";
import Notify from "./utils/Notify";

const headers = [
  {
    key: "id",
    header: "ID",
  },
  {
    key: "username",
    header: "User Name",
  },
  {
    key: "firstname",
    header: "First Name",
  },
  {
    key: "lastname",
    header: "Last Name",
  },
  {
    key: "email",
    header: "Email",
  },
  {
    key: "groups",
    header: "Groups",
  },
  {
    key: "actions",
    header: "Actions",
  },
];

const Users = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const [selectedUser, setSelectedUser] = useState(null);
  const [title, setTitle] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [message, setMessage] = useState("");
  const isAdmin = UserService.isAdminUser();

  const fetchData = async () => {
    let data = await allUsers();
    setRows(data?.payload);
    setLoading(false);
  };

  useEffect(() => {
    fetchData();
  }, [headers]); // eslint-disable-line react-hooks/exhaustive-deps

  const displayData = clientSearchFilter(searchText, rows);

  const handleResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
  };

  const handleDeleteUser = (user) => {
    setSelectedUser(user);
    setActionProps("DELETE_USER");
  };

  const handleModifyGroups = (user) => {
    setSelectedUser(user);
    setActionProps("MODIFY_GROUPS");
  };

  const renderSkeleton = () => {
    const headerLabels = headers?.map((x) => x?.header);
    return (
      <DataTableSkeleton
        columnCount={headerLabels?.length}
        compact={false}
        headers={headerLabels}
        rowCount={3}
        zebra={false}
      />
    );
  };

  return (
    <>
      <Notify title={title} message={message} nkind={notifyKind} setTitle={setTitle} />
      {actionProps === "DELETE_USER" && selectedUser && (
        <DeleteUser
          user={selectedUser}
          setActionProps={setActionProps}
          response={handleResponse}
          fetchData={fetchData}
        />
      )}
      {loading ? (renderSkeleton()) : (
        <>
          <DataTable rows={displayData} headers={headers} isSortable>
            {({
              rows,
              headers,
              getTableProps,
              getHeaderProps,
              getToolbarProps,
              getTableContainerProps,
            }) => {
              return (
                <TableContainer
                  title={"User Details"}
                  {...getTableContainerProps()}
                >
                  <TableToolbar {...getToolbarProps()}>
                    <TableToolbarSearch
                      persistent={true}
                      // tabIndex={batchActionProps.shouldShowBatchActions ? -1 : 0}
                      onChange={(onInputChange) => {
                        setSearchText(onInputChange.target.value);
                      }}
                      placeholder={"Search"}
                    />
                  </TableToolbar>
                  <Table {...getTableProps()}>
                    <TableHead>
                      <TableRow>
                        {headers.map((header) => (
                          <TableHeader
                            key={header.key}
                            {...getHeaderProps({ header })}
                          >
                            {header.header}
                          </TableHeader>
                        ))}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {rows.map((row) => {
                        const userData = displayData.find(u => u.id === row.id);
                        return (
                          <TableRow key={row.id}>
                            {row.cells.map((cell) => (
                              <TableCell key={cell.id}>
                                {cell.info.header === "groups" ? (
                                  <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                                    {Array.isArray(cell.value) ? (
                                      cell.value.map((group, idx) => (
                                        <Tag key={idx} type="blue" size="md">
                                          {group}
                                        </Tag>
                                      ))
                                    ) : cell.value ? (
                                      <Tag type="blue" size="md">
                                        {cell.value}
                                      </Tag>
                                    ) : null}
                                  </div>
                                ) : cell.info.header === "actions" && isAdmin ? (
                                  <OverflowMenu flipped size="sm" ariaLabel="User actions">
                                    <OverflowMenuItem
                                      itemText="Modify Groups"
                                      onClick={() => handleModifyGroups(userData)}
                                    />
                                    <OverflowMenuItem
                                      itemText="Delete User"
                                      hasDivider
                                      isDelete
                                      onClick={() => handleDeleteUser(userData)}
                                    />
                                  </OverflowMenu>
                                ) : cell.info.header === "actions" ? null : (
                                  cell.value
                                )}
                              </TableCell>
                            ))}
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </TableContainer>
              );
            }}
          </DataTable>
          {<FooterPagination displayData={rows} />}
        </>
      )}
      {actionProps === "MODIFY_GROUPS" && selectedUser && (
        <ModifyUserGroups
          user={selectedUser}
          setActionProps={setActionProps}
          response={handleResponse}
          fetchData={fetchData}
        />
      )}
    </>
  );
};

export default Users;
