import React from "react";
import { Pagination } from '@carbon/react';

const FooterPagination = ({displayData}) => {
  // We are using this class to show count at the bottom of the table
  // Remove it when when pagination is introduced
  const count = displayData?.length;
  const displayDataCount = displayData?.length;

  if (displayData.length === 0) {
    return <></>;
  }

  return (
    <div className='dataTableFooter'>
      <Pagination
        totalItems={displayDataCount}
        pageSize={count}
        pageSizes={[count]}
        itemsPerPageText={('Items per page:')}
        itemRangeText={
          (min, max, total) => `${ min }–${ max } ${(' of')} ${ total } ${('items')}`}
        pageRangeText={
          (_current, total) => `${(' of')} ${ total } ${ total === 1 ? ('page') : ('pages') }`}
      />
    </div>
  );
};

export default FooterPagination;
