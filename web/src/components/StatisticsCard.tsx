interface StatItem {
  label: string;
  value: string | number;
  color?: string;
}

interface StatisticsCardProps {
  stats: StatItem[];
}

export default function StatisticsCard({ stats }: StatisticsCardProps) {
  return (
    <div className="bg-white rounded-lg shadow-md p-4">
      <div className="flex flex-col sm:flex-row gap-4 sm:gap-8 justify-center text-center">
        {stats.map((stat, index) => (
          <div key={index} className="flex flex-col">
            <span className="text-sm font-medium text-gray-600">{stat.label}</span>
            <span className={`text-xl font-bold ${stat.color || 'text-blue-600'}`}>
              {stat.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
