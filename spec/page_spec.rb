require_relative "../src/page"

RSpec.describe Page do
  subject(:page) { described_class }

  let(:expected_members) do
    [
      :public?,
      :uid,
      :short_uid,
      :slug,
      :tags,
      :published_at,
      :body,
      :title,
      :hide_from_toc?,
      :url,
      :local_path,
      :public_path,
      :template,
      :layout
    ]
  end

  it { expect(page.members).to eq(expected_members) }
end
