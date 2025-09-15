import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { profileAPI, ProfileUpdateRequest } from '../api/profile';
import LanguageSelector from '../components/LanguageSelector';
import i18n from '../i18n';

// Helper function to convert backend language codes to i18n format
const convertLanguageCode = (backendCode: string): string => {
  return backendCode.replace('_', '-');
};

export default function Profile() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState<ProfileUpdateRequest>({
    first_name: '',
    last_name: '',
    email: '',
    age: undefined,
    height: undefined,
    weight: undefined,
    language: 'en_US',
  });
  const [originalData, setOriginalData] = useState<ProfileUpdateRequest | null>(null);
  const [showUnsavedDialog, setShowUnsavedDialog] = useState(false);
  const [pendingNavigation, setPendingNavigation] = useState<(() => void) | null>(null);
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  const formRef = useRef<HTMLFormElement>(null);

  // Fetch profile data
  const { data: profile, isLoading: profileLoading, error: profileError } = useQuery({
    queryKey: ['profile'],
    queryFn: profileAPI.getProfile,
  });


  // Update mutation
  const updateMutation = useMutation({
    mutationFn: profileAPI.updateProfile,
    onSuccess: (data) => {
      queryClient.setQueryData(['profile'], data);

      // Update user query data as well to reflect language change
      queryClient.setQueryData(['user'], (oldUserData: any) => {
        if (oldUserData) {
          return { ...oldUserData, language: data.language };
        }
        return oldUserData;
      });

      const updatedFormData = {
        first_name: data.first_name || '',
        last_name: data.last_name || '',
        email: data.email,
        age: data.age || undefined,
        height: data.height || undefined,
        weight: data.weight || undefined,
        language: data.language,
      };
      setFormData(updatedFormData);
      setOriginalData(updatedFormData);

      // Apply language change immediately to i18n
      const i18nLangCode = convertLanguageCode(data.language);
      if (i18n.language !== i18nLangCode) {
        i18n.changeLanguage(i18nLangCode);
      }

      // Show success notification
      setNotification({ type: 'success', message: t('profile.saveSuccess') });
    },
    onError: () => {
      // Show error notification
      setNotification({ type: 'error', message: t('profile.saveError') });
    },
  });

  // Set form data when profile is loaded
  useEffect(() => {
    if (profile) {
      const profileData = {
        first_name: profile.first_name || '',
        last_name: profile.last_name || '',
        email: profile.email,
        age: profile.age || undefined,
        height: profile.height || undefined,
        weight: profile.weight || undefined,
        language: profile.language,
      };
      setFormData(profileData);
      setOriginalData(profileData);
    }
  }, [profile]);

  // Check for unsaved changes
  const hasUnsavedChanges = () => {
    if (!originalData) return false;
    return JSON.stringify(formData) !== JSON.stringify(originalData);
  };

  // Handle form changes
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? (value ? Number(value) : undefined) : value,
    }));
  };

  const handleLanguageChange = (language: string) => {
    setFormData(prev => ({
      ...prev,
      language,
    }));
  };

  // Handle save
  const handleSave = async (e?: React.FormEvent) => {
    e?.preventDefault();
    if (!formRef.current?.reportValidity()) {
      return;
    }
    updateMutation.mutate(formData);
  };

  // Handle unsaved changes dialog
  const handleSaveChanges = async () => {
    await handleSave();
    setShowUnsavedDialog(false);
    if (pendingNavigation) {
      pendingNavigation();
      setPendingNavigation(null);
    }
  };

  const handleDiscardChanges = () => {
    setShowUnsavedDialog(false);
    if (pendingNavigation) {
      pendingNavigation();
      setPendingNavigation(null);
    }
  };

  const handleCancelDialog = () => {
    setShowUnsavedDialog(false);
    setPendingNavigation(null);
  };

  // Intercept navigation attempts
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedChanges()) {
        e.preventDefault();
        e.returnValue = '';
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  });


  return (
    <div className="page p-2">

      {profileError && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          Error loading profile: {profileError.message}
        </div>
      )}

      <form ref={formRef} onSubmit={handleSave} className="space-y-6">
        {/* Information Group */}
        <div className="card p-lg">
          <h3 className="text-lg font-semibold mb-4">{t('profile.information')}</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="first_name" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.firstName')}
              </label>
              <input
                type="text"
                id="first_name"
                name="first_name"
                value={formData.first_name}
                onChange={handleInputChange}
                disabled={profileLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
              />
            </div>

            <div>
              <label htmlFor="last_name" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.lastName')}
              </label>
              <input
                type="text"
                id="last_name"
                name="last_name"
                value={formData.last_name}
                onChange={handleInputChange}
                disabled={profileLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
              />
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.email')} *
              </label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleInputChange}
                required
                disabled={profileLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
              />
            </div>

            <div>
              <label htmlFor="language" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.language')}
              </label>
              <LanguageSelector
                value={formData.language}
                onChange={handleLanguageChange}
                className="w-full"
                disabled={profileLoading}
              />
            </div>
          </div>
        </div>

        {/* Physical Parameters Group */}
        <div className="card pl-lg pr-lg">
          <h3 className="text-lg font-semibold mb-4">{t('profile.physicalParameters')}</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label htmlFor="age" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.age')}
              </label>
              <input
                type="number"
                id="age"
                name="age"
                value={formData.age || ''}
                onChange={handleInputChange}
                min="1"
                max="150"
                disabled={profileLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
              />
            </div>

            <div>
              <label htmlFor="weight" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.weight')}
              </label>
              <input
                type="number"
                id="weight"
                name="weight"
                value={formData.weight || ''}
                onChange={handleInputChange}
                min="10"
                max="500"
                step="0.1"
                disabled={profileLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
              />
            </div>

            <div>
              <label htmlFor="height" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.height')}
              </label>
              <input
                type="number"
                id="height"
                name="height"
                value={formData.height || ''}
                onChange={handleInputChange}
                min="50"
                max="300"
                step="0.1"
                disabled={profileLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
              />
            </div>
          </div>
        </div>

        {/* Save Button */}
        <div className="card pl-lg pr-lg flex justify-end">
          <button
            type="submit"
            disabled={profileLoading || updateMutation.isPending || !hasUnsavedChanges()}
            className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {updateMutation.isPending ? t('profile.saving') : t('profile.save')}
          </button>
        </div>
      </form>

      {/* Notification Popup */}
      {notification && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animate-fadeIn px-4">
          <div className="bg-white rounded-lg w-full md:w-[600px] lg:w-[700px] p-8 shadow-xl animate-slideUp text-center">
            <h3 className={`mt-0 mb-4 text-lg font-medium ${
              notification.type === 'success' ? 'text-green-600' : 'text-red-600'
            }`}>
              {notification.type === 'success' ? t('common.success') : t('common.error')}
            </h3>
            <p className="mb-6 text-gray-700">{notification.message}</p>
            <div className="flex justify-center">
              <button
                onClick={() => setNotification(null)}
                className="btn-primary px-4 py-2 text-sm font-medium"
              >
                {t('common.ok')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Unsaved Changes Dialog */}
      {showUnsavedDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold mb-4">{t('profile.unsavedChanges')}</h3>
            <p className="text-gray-600 mb-6">{t('profile.unsavedChangesMessage')}</p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={handleCancelDialog}
                className="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleDiscardChanges}
                className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
              >
                {t('profile.discardChanges')}
              </button>
              <button
                onClick={handleSaveChanges}
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors"
              >
                {t('profile.saveChanges')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}