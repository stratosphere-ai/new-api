import React, { useMemo } from 'react';
import { Table } from '@douyinfe/semi-ui';
import { IllustrationNoResult } from '@douyinfe/semi-illustrations';
import { getAgentKeysColumns } from './AgentKeysColumnDefs';

const AgentKeysTable = ({
  keys,
  loading,
  rowSelection,
  handleRow,
  copyText,
  setEditingKey,
  setShowEdit,
  deleteAgentKey,
  t,
}) => {
  const columns = useMemo(
    () =>
      getAgentKeysColumns({
        t,
        copyText,
        setEditingKey,
        setShowEdit,
        deleteAgentKey,
      }),
    [t, copyText, setEditingKey, setShowEdit, deleteAgentKey],
  );

  return (
    <Table
      columns={columns}
      dataSource={keys}
      rowKey='id'
      loading={loading}
      rowSelection={rowSelection}
      onRow={handleRow}
      pagination={false}
      scroll={{ x: 'max-content' }}
      empty={
        <div style={{ padding: 40, textAlign: 'center' }}>
          <IllustrationNoResult style={{ width: 150, height: 150 }} />
          <p>{t('暂无数据')}</p>
        </div>
      }
    />
  );
};

export default AgentKeysTable;
