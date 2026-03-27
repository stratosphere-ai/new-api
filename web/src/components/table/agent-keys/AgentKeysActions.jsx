import React from 'react';
import { Button } from '@douyinfe/semi-ui';

const AgentKeysActions = ({ setEditingKey, setShowEdit, t }) => {
  return (
    <div className='flex flex-wrap gap-2 w-full md:w-auto order-2 md:order-1'>
      <Button
        type='primary'
        className='flex-1 md:flex-initial'
        onClick={() => {
          setEditingKey({ id: undefined });
          setShowEdit(true);
        }}
        size='small'
      >
        {t('创建 Agent 密钥')}
      </Button>
    </div>
  );
};

export default AgentKeysActions;
