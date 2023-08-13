require_relative "../src/page"

RSpec.describe Page do
  subject(:page) { described_class }

  let(:expected_members) do
    [
      :uid,
      :short_uid,
      :slug,
      :tags,
      :published_at,
      :body,
      :title,
      :url,
      :local_path,
      :public_path,
      :template,
      :layout
    ]
  end

  it { expect(page.members).to eq(expected_members) }
end
