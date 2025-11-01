interface ProfileFormSectionProps {
  title: string;
  children: React.ReactNode;
}

export default function ProfileFormSection({ title, children }: ProfileFormSectionProps) {
  return (
    <div className="bg-white rounded-lg shadow-md p-lg">
      <h3 className="text-lg font-semibold mb-4">{title}</h3>
      {children}
    </div>
  );
}
