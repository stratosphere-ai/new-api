import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { API, copy, showError, showSuccess } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';

export const useAgentKeysData = () => {
  const { t } = useTranslation();

  const [keys, setKeys] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [keyCount, setKeyCount] = useState(0);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);

  const [selectedKeys, setSelectedKeys] = useState([]);

  const [showEdit, setShowEdit] = useState(false);
  const [editingKey, setEditingKey] = useState({ id: undefined });

  const closeEdit = () => {
    setShowEdit(false);
    setTimeout(() => {
      setEditingKey({ id: undefined });
    }, 500);
  };

  const loadKeys = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/agent/keys/');
      const { success, message, data } = res.data;
      if (success) {
        setKeys(data || []);
        setKeyCount((data || []).length);
      } else {
        showError(message);
      }
    } catch (e) {
      showError(e.message);
    }
    setLoading(false);
  };

  const refresh = async () => {
    await loadKeys();
    setSelectedKeys([]);
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制到剪贴板！'));
    } else {
      showError(t('复制失败，请手动复制'));
    }
  };

  const deleteAgentKey = async (id) => {
    setLoading(true);
    try {
      const res = await API.delete(`/api/agent/keys/${id}`);
      if (res.data.success) {
        showSuccess(t('操作成功完成！'));
        await refresh();
      } else {
        showError(res.data.message);
      }
    } catch (e) {
      showError(e.message);
    }
    setLoading(false);
  };

  const handlePageChange = (page) => {
    setActivePage(page);
  };

  const rowSelection = {
    onChange: (selectedRowKeys, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  const handleRow = (record) => {
    if (record.status !== 1) {
      return {
        style: { background: 'var(--semi-color-disabled-border)' },
      };
    }
    return {};
  };

  useEffect(() => {
    loadKeys();
  }, []);

  return {
    keys,
    loading,
    activePage,
    keyCount,
    pageSize,
    setPageSize,
    selectedKeys,
    setSelectedKeys,
    showEdit,
    setShowEdit,
    editingKey,
    setEditingKey,
    closeEdit,
    loadKeys,
    refresh,
    copyText,
    deleteAgentKey,
    handlePageChange,
    rowSelection,
    handleRow,
    t,
  };
};
