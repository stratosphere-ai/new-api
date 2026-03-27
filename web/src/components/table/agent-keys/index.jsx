import React from 'react';
import CardPro from '../../common/ui/CardPro';
import AgentKeysTableComponent from './AgentKeysTable';
import AgentKeysActions from './AgentKeysActions';
import AgentKeysDescription from './AgentKeysDescription';
import EditAgentKeyModal from './modals/EditAgentKeyModal';
import { useAgentKeysData } from '../../../hooks/agent-keys/useAgentKeysData';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { createCardProPagination } from '../../../helpers/utils';

function AgentKeysPage() {
  const data = useAgentKeysData();
  const isMobile = useIsMobile();

  const pagination = createCardProPagination({
    currentPage: data.activePage,
    pageSize: data.pageSize,
    total: data.keyCount,
    onPageChange: data.handlePageChange,
    onPageSizeChange: (size) => data.setPageSize(size),
    isMobile,
    t: data.t,
  });

  return (
    <>
      <EditAgentKeyModal
        visible={data.showEdit}
        editingKey={data.editingKey}
        handleClose={data.closeEdit}
        refresh={data.refresh}
      />

      <CardPro
        type='type1'
        descriptionArea={<AgentKeysDescription t={data.t} />}
        actionsArea={
          <AgentKeysActions
            setEditingKey={data.setEditingKey}
            setShowEdit={data.setShowEdit}
            t={data.t}
          />
        }
        paginationArea={pagination}
      >
        <AgentKeysTableComponent
          keys={data.keys}
          loading={data.loading}
          rowSelection={data.rowSelection}
          handleRow={data.handleRow}
          copyText={data.copyText}
          setEditingKey={data.setEditingKey}
          setShowEdit={data.setShowEdit}
          deleteAgentKey={data.deleteAgentKey}
          t={data.t}
        />
      </CardPro>
    </>
  );
}

export default AgentKeysPage;
