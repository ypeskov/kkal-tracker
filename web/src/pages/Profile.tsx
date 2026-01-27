import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { profileService, ProfileUpdateRequest } from '@/api/profile';
import { weightService } from '@/api/weight';
import { metricsService } from '@/api/metrics';
import LanguageSelector from '@/components/LanguageSelector';
import ProfileFormField from '@/components/ProfileFormField';
import ProfileFormSection from '@/components/ProfileFormSection';
import WeightDisplay from '@/components/WeightDisplay';
import NotificationPopup from '@/components/NotificationPopup';
import UnsavedChangesDialog from '@/components/UnsavedChangesDialog';
import i18n from '@/i18n';

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
    gender: undefined,
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
    queryFn: profileService.getProfile,
  });

  // Fetch latest weight from history
  const { data: weightHistory } = useQuery({
    queryKey: ['latestWeight'],
    queryFn: () => weightService.getWeightHistory(),
  });

  // Fetch health metrics
  const { data: healthMetrics } = useQuery({
    queryKey: ['healthMetrics'],
    queryFn: metricsService.getHealthMetrics,
  });


  // Update mutation
  const updateMutation = useMutation({
    mutationFn: profileService.updateProfile,
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
        gender: data.gender || undefined,
        language: data.language,
      };
      setFormData(updatedFormData);
      setOriginalData(updatedFormData);

      // Invalidate health metrics to recalculate after profile update
      queryClient.invalidateQueries({ queryKey: ['healthMetrics'] });

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
        gender: profile.gender || undefined,
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

  const handleSelectChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value || undefined,
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
    <div className="max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8">

      {profileError && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          Error loading profile: {profileError.message}
        </div>
      )}

      <form ref={formRef} onSubmit={handleSave} className="space-y-6">
        {/* Information Group */}
        <ProfileFormSection title={t('profile.information')}>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <ProfileFormField
              label={t('profile.firstName')}
              id="first_name"
              name="first_name"
              type="text"
              value={formData.first_name}
              onChange={handleInputChange}
              disabled={profileLoading}
            />

            <ProfileFormField
              label={t('profile.lastName')}
              id="last_name"
              name="last_name"
              type="text"
              value={formData.last_name}
              onChange={handleInputChange}
              disabled={profileLoading}
            />

            <ProfileFormField
              label={t('profile.email')}
              id="email"
              name="email"
              type="email"
              value={formData.email}
              onChange={handleInputChange}
              required
              disabled={true}
            />

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
        </ProfileFormSection>

        {/* Physical Parameters Group */}
        <ProfileFormSection title={t('profile.physicalParameters')}>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <ProfileFormField
              label={t('profile.age')}
              id="age"
              name="age"
              type="number"
              value={formData.age}
              onChange={handleInputChange}
              min={1}
              max={150}
              disabled={profileLoading}
            />

            <WeightDisplay weightHistory={weightHistory} />

            <ProfileFormField
              label={t('profile.height')}
              id="height"
              name="height"
              type="number"
              value={formData.height}
              onChange={handleInputChange}
              min={50}
              max={300}
              step={0.1}
              disabled={profileLoading}
            />

            <div>
              <label htmlFor="gender" className="block text-sm font-medium text-gray-700 mb-1">
                {t('profile.gender')}
              </label>
              <select
                id="gender"
                name="gender"
                value={formData.gender || ''}
                onChange={handleSelectChange}
                disabled={profileLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
              >
                <option value="">{t('profile.genderSelect')}</option>
                <option value="male">{t('profile.genderMale')}</option>
                <option value="female">{t('profile.genderFemale')}</option>
              </select>
            </div>
          </div>
        </ProfileFormSection>

        {/* Health Metrics Group */}
        <ProfileFormSection title={t('profile.healthMetrics')}>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* BMI Card */}
            <div className="p-4 bg-blue-50 rounded-lg">
              <div className="text-sm text-gray-600 mb-1">{t('profile.bmi')}</div>
              {healthMetrics?.bmi ? (
                <>
                  <div className="text-2xl font-bold text-blue-600">{healthMetrics.bmi.toFixed(1)}</div>
                  {healthMetrics.bmi_category && (
                    <div className="text-xs text-gray-500 mt-1">
                      {t(`profile.bmiCategory.${healthMetrics.bmi_category}`)}
                    </div>
                  )}
                </>
              ) : (
                <div className="text-sm text-gray-500 mt-2">
                  {t('profile.bmiRequires')}:
                  <ul className="mt-1 ml-4 list-disc text-xs">
                    {!profile?.height && <li>{t('profile.height')}</li>}
                    {!profile?.weight && <li>{t('profile.weight')}</li>}
                  </ul>
                </div>
              )}
            </div>
            
            {/* BMR Card */}
            <div className="p-4 bg-green-50 rounded-lg">
              <div className="text-sm text-gray-600 mb-1">{t('profile.bmr')}</div>
              {healthMetrics?.bmr ? (
                <>
                  <div className="text-2xl font-bold text-green-600">{Math.round(healthMetrics.bmr)}</div>
                  <div className="text-xs text-gray-500 mt-1">{t('common.kcal')}/day</div>
                </>
              ) : (
                <div className="text-sm text-gray-500 mt-2">
                  {t('profile.bmrRequires')}:
                  <ul className="mt-1 ml-4 list-disc text-xs">
                    {!profile?.age && <li>{t('profile.age')}</li>}
                    {!profile?.height && <li>{t('profile.height')}</li>}
                    {!profile?.weight && <li>{t('profile.weight')}</li>}
                    {!profile?.gender && <li>{t('profile.gender')}</li>}
                  </ul>
                </div>
              )}
            </div>

            {/* TDEE Card */}
            <div className="p-4 bg-purple-50 rounded-lg">
              <div className="text-sm text-gray-600 mb-1">{t('profile.tdee')}</div>
              {healthMetrics?.tdee ? (
                <>
                  <div className="text-2xl font-bold text-purple-600">{Math.round(healthMetrics.tdee)}</div>
                  <div className="text-xs text-gray-500 mt-1">{t('common.kcal')}/day (sedentary)</div>
                </>
              ) : (
                <div className="text-sm text-gray-500 mt-2">
                  {t('profile.tdeeRequires')}:
                  <ul className="mt-1 ml-4 list-disc text-xs">
                    {!profile?.age && <li>{t('profile.age')}</li>}
                    {!profile?.height && <li>{t('profile.height')}</li>}
                    {!profile?.weight && <li>{t('profile.weight')}</li>}
                    {!profile?.gender && <li>{t('profile.gender')}</li>}
                  </ul>
                </div>
              )}
            </div>
          </div>
        </ProfileFormSection>

        {/* Save Button */}
        <div className="bg-white rounded-lg shadow-md pl-lg pr-lg flex justify-end">
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
        <NotificationPopup
          type={notification.type}
          message={notification.message}
          onClose={() => setNotification(null)}
        />
      )}

      {/* Unsaved Changes Dialog */}
      {showUnsavedDialog && (
        <UnsavedChangesDialog
          onSave={handleSaveChanges}
          onDiscard={handleDiscardChanges}
          onCancel={handleCancelDialog}
        />
      )}
    </div>
  );
}