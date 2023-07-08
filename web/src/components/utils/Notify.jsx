import React from "react";
import { InlineNotification } from "@carbon/react";

const Notify = ({ title, message, nkind, setTitle }) => {
    return (
        <>
            <div>
                {title && (
                    <InlineNotification
                        kind={nkind}
                        title={title}
                        subtitle={message}
                        onClose={() => {
                            setTitle("");
                        }}
                    />
                )}
            </div>
        </>
    );
};
export default Notify;
