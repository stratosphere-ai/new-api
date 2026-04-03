import React from 'react';
import { Empty, Button } from '@douyinfe/semi-ui';
import { useNavigate } from 'react-router-dom';

const EmptyStateGuide = ({
  icon,
  title,
  description,
  actionLabel,
  actionTo,
  onAction,
}) => {
  const navigate = useNavigate();

  const handleAction = () => {
    if (onAction) {
      onAction();
    } else if (actionTo) {
      navigate(actionTo);
    }
  };

  return (
    <div className='empty-state-container'>
      <Empty
        image={icon}
        title={title}
        description={description}
      />
      {(actionLabel && (actionTo || onAction)) && (
        <Button
          type='primary'
          theme='solid'
          className='!rounded-lg mt-4'
          onClick={handleAction}
        >
          {actionLabel}
        </Button>
      )}
    </div>
  );
};

export default EmptyStateGuide;
