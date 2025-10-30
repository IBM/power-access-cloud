import { ToastNotification } from "@carbon/react";
import './ToastNotify.scss'

const ToastNotify = ({ title, subTitle, kind, caption, setTitle }) => {
  return (
    <div className="toast-notify">
      {title && (
        <ToastNotification
          title={title}
          subtitle={subTitle}
          kind={kind}
          caption={caption}
          onClose={() => setTitle("")}
          timeout={5000}
        />
      )}
    </div>
  );
};

export default ToastNotify;
